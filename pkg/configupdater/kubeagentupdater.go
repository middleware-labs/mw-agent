package configupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/middleware-labs/mw-agent/pkg/agent"
)

const Timestamp string = "timestamp"

type ComponentType int

const (
	Deployment ComponentType = iota
	DaemonSet
)

func (d ComponentType) String() string {
	switch d {
	case Deployment:
		return "deployment"
	case DaemonSet:
		return "daemonset"
	}
	return "unknown"
}

// KubeAgentOptions takes in various options for KubeAgentUpdater
type KubeAgentOptions func(h *KubeAgent)

// NewKubeAgent returns new agent monitor for Kubernetes with given options.
func NewKubeAgent(cfg BaseConfig, agentVersion string, clientset kubernetes.Interface, logger *zap.Logger) (*KubeAgent, error) {
	var agent KubeAgent
	agent.BaseConfig = cfg
	agent.version = agentVersion
	if logger == nil {
		agent.logger, _ = zap.NewProduction()
	} else {
		agent.logger = logger
	}

	duration, err := time.ParseDuration(cfg.ConfigCheckInterval)
	if err != nil {
		return nil, err
	}
	agent.configCheckDuration = duration

	agent.clientset = clientset

	return &agent, nil
}

// ListenForConfigChanges listens for configuration changes for the
// agent on the Middleware backend and restarts the agent if configuration
// has changed.
func (c *KubeAgent) ListenForConfigChanges(ctx context.Context, errCh chan<- error,
	stopCh <-chan struct{}) error {

	errCh <- c.callRestartStatusAPI(ctx, true)
	ticker := time.NewTicker(c.configCheckDuration)

	for {
		c.logger.Info("checking for config change every",
			zap.String("restartInterval", c.configCheckDuration.String()))
		select {
		case <-stopCh:
			ticker.Stop()
			return nil
		case <-ticker.C:
			errCh <- c.callRestartStatusAPI(ctx, false)
		}
	}
}

// callRestartStatusAPI checks if there is an update in the otel-config at Middleware Backend
// For a particular account
func (c *KubeAgent) callRestartStatusAPI(ctx context.Context, first bool) error {
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}

	baseURL := u.JoinPath(apiPathForRestart)
	baseURL = baseURL.JoinPath(c.APIKey)
	params := url.Values{}
	params.Add("platform", "k8s")
	params.Add("host_id", c.ClusterName)
	params.Add("cluster", c.ClusterName)
	params.Add("agent_version", c.version)

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters
	url := baseURL.String()
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to call restart api for url %s: %w",
			url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("restart api returned non-200 status: %d", resp.StatusCode)
	}

	var apiResponse agent.ApiResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal restart api response: %w", err)
	}

	c.HandleIntegrationHealthChecks(apiResponse.Response)

	if apiResponse.Rollout.Daemonset || first {
		c.logger.Info("redeploying mw-agent daemonset")
		updateConfigMapErr := c.UpdateConfigMap(ctx, DaemonSet)
		if updateConfigMapErr != nil {
			return updateConfigMapErr
		}

		if err := c.restartKubeAgent(ctx, DaemonSet); err != nil {
			return fmt.Errorf("error getting updated config: %w", err)
		}
	}

	if apiResponse.Rollout.Deployment || first {
		c.logger.Info("redeploying mw-agent deployment")
		updateConfigMapErr := c.UpdateConfigMap(ctx, Deployment)
		if updateConfigMapErr != nil {
			return updateConfigMapErr
		}

		if err := c.restartKubeAgent(ctx, Deployment); err != nil {
			return fmt.Errorf("error restarting mw-agent: %w", err)
		}
	}

	return err
}

// restartKubeAgent rollout restarts agent's data scraping components
func (c *KubeAgent) restartKubeAgent(ctx context.Context, componentType ComponentType) error {
	return c.rolloutRestart(ctx, componentType)
}

// rolloutRestart reloads the k8s components
// based on component type
func (c *KubeAgent) rolloutRestart(ctx context.Context, componentType ComponentType) error {

	switch componentType {
	case DaemonSet:
		daemonSet, err := c.clientset.AppsV1().DaemonSets(c.AgentNamespaceName).Get(ctx, c.DaemonsetName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		daemonSet.Spec.Template.ObjectMeta.Labels[Timestamp] = fmt.Sprintf("%d", metav1.Now().Unix())
		_, err = c.clientset.AppsV1().DaemonSets(c.AgentNamespaceName).Update(ctx, daemonSet, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

	case Deployment:
		deployment, err := c.clientset.AppsV1().Deployments(c.AgentNamespaceName).Get(ctx, c.DeploymentName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Template.ObjectMeta.Labels[Timestamp] = fmt.Sprintf("%d", metav1.Now().Unix())
		_, err = c.clientset.AppsV1().Deployments(c.AgentNamespaceName).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// updateConfigMap gets the latest configmap from Middleware backend and updates the k8s configmap
// based on component type
func (c *KubeAgent) UpdateConfigMap(ctx context.Context, componentType ComponentType) error {
	fmt.Println(c)
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}

	baseURL := u.JoinPath(apiPathForYAML)
	baseURL = baseURL.JoinPath(c.APIKey)
	params := url.Values{}
	params.Add("platform", "k8s")
	params.Add("component_type", componentType.String())
	params.Add("host_id", c.ClusterName)
	params.Add("cluster", c.ClusterName)
	params.Add("agent_version", c.version)

	if c.BaseConfig.EnableDataDogReceiver {
		params.Add("enable_datadog_receiver", "true")
	}

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	resp, err := http.Get(baseURL.String())
	if err != nil {
		c.logger.Error("failed to call Restart-API", zap.String("url", baseURL.String()), zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get configuration api returned non-200 status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal JSON response into ApiResponse struct
	var apiResponse apiResponseForYAML
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal api response: %w", err)
	}

	// Verify API Response
	if !apiResponse.Status {
		return fmt.Errorf("failure status from api response for ingestion rules: %t",
			apiResponse.Status)
	}

	var apiYAMLConfig map[string]interface{}
	if len(apiResponse.Config.DaemonSet) == 0 && len(apiResponse.Config.Deployment) == 0 {
		return fmt.Errorf("failed to get valid response, config docker len: %d, config no docker len: %d",
			len(apiResponse.Config.Docker), len(apiResponse.Config.NoDocker))
	}

	apiYAMLConfig = apiResponse.Config.Deployment
	if componentType == DaemonSet {
		apiYAMLConfig = apiResponse.Config.DaemonSet
	}

	yamlData, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal api data: %w", err)
	}

	switch componentType {
	case DaemonSet:
		// Retrieve the existing ConfigMap
		existingDaemonsetConfigMap, err := c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Get(ctx, c.DaemonsetConfigMapName, metav1.GetOptions{})

		if err != nil {
			if k8serrors.IsNotFound(err) {
				existingDaemonsetConfigMap, err = c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Create(ctx, &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: c.DaemonsetConfigMapName,
					},
					Data: map[string]string{
						"otel-config": string(yamlData),
					},
				}, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("failed to create daemonset configmap: %w", err)
				}
			} else {
				return fmt.Errorf("failed to get daemonset configmap: %w", err)
			}
		} else {
			existingDaemonsetConfigMap.Data["otel-config"] = string(yamlData)
		}

		// Update the ConfigMap
		updatedConfigMap, err := c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Update(ctx, existingDaemonsetConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update daemonset configmap: %w", err)
		}

		c.logger.Info("Daemonset configmap updated successfully ", zap.String("configmap", updatedConfigMap.Name))
	case Deployment:
		// Retrieve the existing ConfigMap
		existingDeploymentConfigMap, err := c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Get(ctx, c.DeploymentConfigMapName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				existingDeploymentConfigMap, err = c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Create(ctx, &v1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Name: c.DeploymentConfigMapName,
					},
					Data: map[string]string{
						"otel-config": string(yamlData),
					},
				}, metav1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("failed to create deployment configmap: %w", err)
				}
			} else {
				return fmt.Errorf("failed to get deployment configmap: %w", err)
			}
		} else {
			existingDeploymentConfigMap.Data["otel-config"] = string(yamlData)
		}

		// Update the ConfigMap
		updatedDeploymentConfigMap, err := c.clientset.CoreV1().ConfigMaps(c.AgentNamespaceName).Update(ctx, existingDeploymentConfigMap, metav1.UpdateOptions{})
		if err != nil || updatedDeploymentConfigMap == nil {
			return fmt.Errorf("failed to update deployment configmap: %w", err)
		}

		c.logger.Info("Deployment configmap updated successfully", zap.String("configmap", updatedDeploymentConfigMap.Name))
	}

	return nil

}

func (c *KubeAgent) HandleIntegrationHealthChecks(setting agent.AgentSettingModels) error {
	meta := setting.MetaData
	if meta == nil {
		return fmt.Errorf("metadata missing in API response")
	}

	platformMeta, ok := meta["k8s"].(map[string]interface{})
	if !ok {
		err := fmt.Errorf("metadata missing for platform: %s", "k8s")
		c.logger.Error("metadata missing for platform", zap.String("platform", "k8s"))
		return err
	}

	// Fetch config for the platform for extracting settings
	cfgPlatform, ok := setting.Config["k8s"]
	if !ok {
		err := fmt.Errorf("config missing for platform: %s", "k8s")
		c.logger.Error("config missing for platform", zap.String("platform", "k8s"))
		return err
	}

	// Iterate all integrations: rabbitmq_config, mysql_config, etc.
	for integrationKey, rawCfg := range platformMeta {
		integrationCfg, ok := rawCfg.(map[string]interface{})
		if !ok {
			continue
		}

		shouldTest, ok := integrationCfg["should_test"].(bool)
		if !ok || !shouldTest {
			continue
		}

		// Extract settings from config (NOT metadata)
		cfgEntryRaw, ok := cfgPlatform[integrationKey].(map[string]interface{})
		if !ok {
			err := fmt.Errorf("config entry missing for integration: %s", integrationKey)
			c.logger.Error("config entry missing for integration", zap.String("integration", integrationKey))
			return err
		}

		settingsArr, ok := cfgEntryRaw["settings"].([]interface{})
		if !ok || len(settingsArr) == 0 {
			err := fmt.Errorf("%s.settings missing or empty", integrationKey)
			c.logger.Error("integration settings missing or empty", zap.String("integration", integrationKey))
			return err
		}

		// GET FIRST ENTRY
		settings, ok := settingsArr[0].(map[string]interface{})
		if !ok {
			err := fmt.Errorf("%s invalid settings format", integrationKey)
			c.logger.Error("invalid settings format", zap.String("integration", integrationKey))
			return err
		}

		// Supported integrations
		switch integrationKey {
		case "rabbitmq_config":
			if err := c.testRabbitMQFromMetadata(integrationKey, settings); err != nil {
				c.logger.Error("RabbitMQ health check failed", zap.String("integration", integrationKey))
				return err
			}
		default:
			c.logger.Warn("Unsupported integration", zap.String("integration", integrationKey))
			continue
		}
	}

	return nil
}

func (c *KubeAgent) testRabbitMQFromMetadata(
	integrationKey string,
	settings map[string]interface{},
) error {
	// --- STEP 1: Perform aliveness test ---
	isAlive, err := c.testRabbitMQConnection(settings)
	if err != nil {
		c.logger.Error("RabbitMQ aliveness test failed",
			zap.String("integration", integrationKey),
			zap.Error(err),
		)
		return err
	}

	c.logger.Info("RabbitMQ aliveness test completed",
		zap.Bool("alive", isAlive),
		zap.String("integration", integrationKey),
	)

	// --- STEP 2: Update server with status ---
	if err := c.updateRabbitMQHealthStatus(integrationKey, isAlive); err != nil {
		c.logger.Error("Failed to update health check status",
			zap.String("integration", integrationKey),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (c *KubeAgent) testRabbitMQConnection(settings map[string]interface{}) (bool, error) {
	username, _ := settings["username"].(string)
	password, _ := settings["password"].(string)
	endpoint, _ := settings["endpoint"].(string)
	if username == "" || password == "" || endpoint == "" {
		err := fmt.Errorf("rabbitmq credentials missing")
		c.logger.Error("Missing RabbitMQ credentials", zap.Error(err))
		return false, err
	}

	// If endpoint does not start with http:// or https:// then add http://
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	// Build the final API URL
	apiURL := fmt.Sprintf("%s/%s", endpoint, agent.ApiRabbitMQAliveness)
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("error creating aliveness request: %w", err)
	}
	req.SetBasicAuth(username, password)

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error calling aliveness API: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return false, fmt.Errorf("invalid JSON from RabbitMQ: %w", err)
	}

	status, _ := data["status"].(string)
	isAlive := status == "ok"

	return isAlive, nil
}

func (c *KubeAgent) updateRabbitMQHealthStatus(integrationKey string, isAlive bool) error {
	u, err := url.Parse(c.Target)
	if err != nil {
		return err
	}

	// Remove port if present
	host := u.Hostname()

	// Build final base URL without port
	baseURL := fmt.Sprintf("%s://%s", u.Scheme, host)

	client := &http.Client{Timeout: 10 * time.Second}
	updateURL := fmt.Sprintf(
		"%s/%s/%s/%s",
		baseURL,
		agent.ApiAgentHealthCheckUpdate,
		c.APIKey,
		c.ClusterName)

	reqBody := agent.HealthCheckRequest{
		Platform:         "k8s",
		IntegrationKey:   integrationKey,
		ShouldTest:       false,
		IsConnectionLive: isAlive,
	}

	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("PUT", updateURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create PUT request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("ApiKey", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("update request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Warn("RabbitMQ update API returned non-200",
			zap.Int("status", resp.StatusCode),
			zap.String("integration", integrationKey),
		)
		return fmt.Errorf("update API returned non-200: %d", resp.StatusCode)
	}

	c.logger.Info("RabbitMQ health check update successful",
		zap.String("integration", integrationKey),
		zap.String("clusterName", c.ClusterName),
	)

	return nil
}
