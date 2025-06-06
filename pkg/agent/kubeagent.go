package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	yaml "gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/logdedupprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/probabilisticsamplerprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/redactionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/datadogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/journaldreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sobjectsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbatlasreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/nginxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/rabbitmqreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/sqlserverreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/debugexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const Timestamp string = "timestamp"

// KubeAgent implements Agent interface for Kubernetes
type KubeAgent struct {
	KubeConfig
	logger *zap.Logger
}

type KubeAgentMonitor struct {
	Clientset kubernetes.Interface
	KubeAgentMonitorConfig
	KubeConfig
	ClusterName string
	logger      *zap.Logger
	Version     string
}

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

// KubeOptions takes in various options for KubeAgent
type KubeOptions func(h *KubeAgent)

// KubeAgentMonitorOptions takes in various options for KubeAgentMonitor
type KubeAgentMonitorOptions func(h *KubeAgentMonitor)

// WithKubeAgentLogger sets the logger to be used with agent logs
func WithKubeAgentLogger(logger *zap.Logger) KubeOptions {
	return func(h *KubeAgent) {
		h.logger = logger
	}
}

// NewKubeAgent returns new agent for Kubernetes with given options.
func NewKubeAgent(cfg KubeConfig, opts ...KubeOptions) *KubeAgent {
	var agent KubeAgent
	agent.KubeConfig = cfg
	for _, apply := range opts {
		apply(&agent)
	}

	if agent.logger == nil {
		agent.logger, _ = zap.NewProduction()
	}

	return &agent
}

// NewKubeAgentMonitor returns new agent monitor for Kubernetes with given options.
func NewKubeAgentMonitor(cfg KubeConfig, opts ...KubeAgentMonitorOptions) *KubeAgentMonitor {
	var agent KubeAgentMonitor
	agent.KubeConfig = cfg
	for _, apply := range opts {
		apply(&agent)
	}

	if agent.logger == nil {
		agent.logger, _ = zap.NewProduction()
	}

	return &agent
}

// GetFactories get otel factories for KubeAgent
func (k *KubeAgent) GetFactories(_ context.Context) (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}
	factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
	// frontend.NewAuthFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Receivers, err = receiver.MakeFactoryMap([]receiver.Factory{
		otlpreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		k8sclusterreceiver.NewFactory(),
		k8seventsreceiver.NewFactory(),
		kubeletstatsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		k8sobjectsreceiver.NewFactory(),
		statsdreceiver.NewFactory(),
		journaldreceiver.NewFactory(),
		rabbitmqreceiver.NewFactory(),
		sqlserverreceiver.NewFactory(),
		nginxreceiver.NewFactory(),
		mongodbatlasreceiver.NewFactory(),
		datadogreceiver.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = exporter.MakeFactoryMap([]exporter.Factory{
		debugexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		kafkaexporter.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = processor.MakeFactoryMap([]processor.Factory{
		// frontend.NewProcessorFactory(),
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		k8sattributesprocessor.NewFactory(),
		deltatorateprocessor.NewFactory(),
		cumulativetodeltaprocessor.NewFactory(),
		metricstransformprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		groupbyattrsprocessor.NewFactory(),
		logdedupprocessor.NewFactory(),
		probabilisticsamplerprocessor.NewFactory(),
		redactionprocessor.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

// ListenForConfigChanges listens for configuration changes for the
// agent on the Middleware backend and restarts the agent if configuration
// has changed.
func (c *KubeAgentMonitor) ListenForKubeOtelConfigChanges(ctx context.Context) error {
	err := c.callRestartStatusAPI(ctx)
	if err != nil {
		c.logger.Info("error restarting agent on config change",
			zap.Error(err))
	}
	return nil
}

// callRestartStatusAPI checks if there is an update in the otel-config at Middleware Backend
// For a particular account
func (c *KubeAgentMonitor) callRestartStatusAPI(ctx context.Context) error {

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
	params.Add("agent_version", c.Version)

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

	if apiResponse.Rollout.Daemonset {
		c.logger.Info("restarting mw-agent")
		if err := c.restartKubeAgent(ctx, DaemonSet); err != nil {
			return fmt.Errorf("error getting updated config: %w", err)
		}
	}

	if apiResponse.Rollout.Deployment {
		c.logger.Info("restarting mw-agent")
		if err := c.restartKubeAgent(ctx, Deployment); err != nil {
			return fmt.Errorf("error restarting mw-agent: %w", err)
		}
	}

	return err
}

// restartKubeAgent rewrites the configmaps and rollout restarts agent's data scraping components
func (c *KubeAgentMonitor) restartKubeAgent(ctx context.Context, componentType ComponentType) error {

	updateConfigMapErr := c.UpdateConfigMap(ctx, componentType)
	if updateConfigMapErr != nil {
		return updateConfigMapErr
	}

	return c.rolloutRestart(ctx, componentType)

}

func (c *KubeAgentMonitor) SetClientSet() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	c.Clientset = clientset
	return nil
}

// rolloutRestart reloads the k8s components
// based on component type
func (c *KubeAgentMonitor) rolloutRestart(ctx context.Context, componentType ComponentType) error {

	switch componentType {
	case DaemonSet:
		daemonSet, err := c.Clientset.AppsV1().DaemonSets(c.AgentNamespace).Get(ctx, c.Daemonset, metav1.GetOptions{})
		if err != nil {
			return err
		}

		daemonSet.Spec.Template.ObjectMeta.Labels[Timestamp] = fmt.Sprintf("%d", metav1.Now().Unix())
		_, err = c.Clientset.AppsV1().DaemonSets(c.AgentNamespace).Update(ctx, daemonSet, metav1.UpdateOptions{})
		if err != nil {
			return err
		}

	case Deployment:
		deployment, err := c.Clientset.AppsV1().Deployments(c.AgentNamespace).Get(ctx, c.Deployment, metav1.GetOptions{})
		if err != nil {
			return err
		}

		deployment.Spec.Template.ObjectMeta.Labels[Timestamp] = fmt.Sprintf("%d", metav1.Now().Unix())
		_, err = c.Clientset.AppsV1().Deployments(c.AgentNamespace).Update(ctx, deployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// updateConfigMap gets the latest configmap from Middleware backend and updates the k8s configmap
// based on component type
func (c *KubeAgentMonitor) UpdateConfigMap(ctx context.Context, componentType ComponentType) error {

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
	params.Add("agent_version", c.Version)

	if c.EnableDataDogReceiver {
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
		existingDaemonsetConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Get(ctx, c.DaemonsetConfigMap, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get configmap: %w", err)
		}

		// Modify the content of the ConfigMap
		existingDaemonsetConfigMap.Data["otel-config"] = string(yamlData)

		// Update the ConfigMap
		updatedConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Update(ctx, existingDaemonsetConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update configmap: %w", err)
		}

		c.logger.Info("ConfigMap updated successfully ", zap.String("configmap", updatedConfigMap.Name))
	case Deployment:
		// Retrieve the existing ConfigMap
		existingDeploymentConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Get(ctx, c.DeploymentConfigMap, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get configmap: %w", err)
		}

		// Modify the content of the ConfigMap
		existingDeploymentConfigMap.Data["otel-config"] = string(yamlData)

		// Update the ConfigMap
		updatedDeploymentConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Update(ctx, existingDeploymentConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update configmap: %w", err)
		}

		c.logger.Info("ConfigMap updated successfully", zap.String("configmap", updatedDeploymentConfigMap.Name))
	}

	return nil

}
