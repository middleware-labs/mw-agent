package configupdater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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
			err := c.callRestartStatusAPI(ctx, false)
			if err == nil {
				// If restart check succeeded, try to apply config class (only once)
				c.applyConfigOnce.Do(func() {
					if applyErr := c.applyConfigClassToCluster(); applyErr != nil {
						c.logger.Error("failed to apply config class", zap.Error(applyErr))
					}
				})
			}
			errCh <- err
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

	var apiResponse apiResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal restart api response: %w", err)
	}

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

// applyConfigClassToCluster applies config class to the Kubernetes cluster
func (c *KubeAgent) applyConfigClassToCluster() error {
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}

	// Build the URL: /agent/public/setting/config-groups/{groupName}/{token}
	baseURL := u.JoinPath(apiPathForConfigGroups)
	baseURL = baseURL.JoinPath(c.APIKey)
	baseURL = baseURL.JoinPath("default")

	// Prepare request body
	reqBody := map[string][]string{
		"hostIds": {c.ClusterName},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest(http.MethodPut, baseURL.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call config groups api for url %s: %w", baseURL.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("config groups api returned non-200 status: %d", resp.StatusCode)
	}

	var apiResponse struct {
		Status  bool        `json:"status"`
		Message string      `json:"message"`
		Setting interface{} `json:"setting"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal config groups api response: %w", err)
	}

	if !apiResponse.Status {
		return fmt.Errorf("config groups api returned error: %s", apiResponse.Message)
	}

	c.logger.Info("successfully applied config class to cluster",
		zap.String("group_name", "default"),
		zap.String("cluster_id", c.ClusterName))

	return nil
}
