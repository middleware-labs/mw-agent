package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	yaml "gopkg.in/yaml.v2"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8seventsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kubeletstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type Response struct {
	Data []AccountData `json:"data"`
}

type AccountData struct {
	AccountID string `json:"accountId"`
	Region    string `json:"region"`
	Namespace string `json:"namespace"`
}

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
	//healthcheckextension.NewFactory(),
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
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = exporter.MakeFactoryMap([]exporter.Factory{
		loggingexporter.NewFactory(),
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
	fmt.Println("Listening...")
	c.callRestartStatusAPI(ctx)
	return nil
}

// divideData divides data into n almost equal parts
func divideData(data []AccountData, n int) [][]AccountData {
	fmt.Println("dividedata....")
	var divided [][]AccountData
	chunkSize := (len(data) + n - 1) / n

	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		divided = append(divided, data[i:end])
	}

	return divided
}

// callRestartStatusAPI checks if there is an update in the otel-config at Middleware Backend
// For a particular account
func (c *KubeAgentMonitor) callRestartStatusAPI(ctx context.Context) {
	fmt.Println("Restarting...")
	// Create a shared informer factory for HPA resources
	factory := informers.NewSharedInformerFactory(c.Clientset, time.Minute)
	hpaInformer := factory.Autoscaling().V1().HorizontalPodAutoscalers().Informer()

	stopCh := make(chan struct{})
	defer close(stopCh)

	hpaInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldHPA := oldObj.(*autoscalingv1.HorizontalPodAutoscaler)
			newHPA := newObj.(*autoscalingv1.HorizontalPodAutoscaler)
			if oldHPA.Status.CurrentReplicas != newHPA.Status.CurrentReplicas {
				fmt.Printf("HPA updated: %s/%s, replicas: %d -> %d\n", newHPA.Namespace, newHPA.Name, oldHPA.Status.CurrentReplicas, newHPA.Status.CurrentReplicas)

				resp, err := http.Get(os.Getenv("MW_AWS_JOBS_LIST_URL"))
				if err != nil {
					fmt.Printf("Error making GET request: %v\n", err)
					return
				}
				defer resp.Body.Close()

				// Read the response body
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Error reading response body: %v\n", err)
					return
				}

				// Parse the JSON response into the custom struct
				var response Response
				err = json.Unmarshal(body, &response)
				if err != nil {
					fmt.Printf("Error unmarshalling JSON response: %v\n", err)
					return
				}

				// Print the parsed data
				for _, data := range response.Data {
					fmt.Printf("AccountID: %s, Region: %s, Namespace: %s\n", data.AccountID, data.Region, data.Namespace)
				}

				// Divide the data into N almost equal parts
				dividedData := divideData(response.Data, int(newHPA.Status.CurrentReplicas))

				// Create receivers for each part
				for i, part := range dividedData {

					// todo: create / delete configmaps based on new replicas, connect them with statefulset pods

					fmt.Printf("Receiver %d:\n", i+1)
					for _, data := range part {
						fmt.Printf("awscloudwatchmetrics/%s_%s_%s:\n", data.AccountID, data.Region, data.Namespace)
						fmt.Printf("    region: %s\n", data.Region)
						fmt.Printf("    profile: \"default\"\n")
						fmt.Printf("    imds_endpoint: \"\"\n")
						fmt.Printf("    poll_interval: \"5m\"\n")
						fmt.Printf("    metrics:\n")
						fmt.Printf("      named:\n")
						fmt.Printf("      - namespace: \"AWS/%s\"\n", data.Namespace)
						fmt.Printf("        metric_name: \"CPUUtilization\"\n")
						fmt.Printf("        period: \"5m\"\n")
						fmt.Printf("        aws_aggregation: \"Sum\"\n")
					}
				}
			}

		},
	})

	// Start the informer
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	// Block forever
	<-stopCh
	// if apiResponse.Rollout.Deployment {
	// 	c.logger.Info("restarting mw-agent")
	// 	if err := c.restartKubeAgent(ctx, Deployment); err != nil {
	// 		c.logger.Error("error restarting mw-agent deployment", zap.Error(err))
	// 		return err
	// 	}
	// }

	// return err
}

// restartKubeAgent rewrites the configmaps and rollout restarts agent's data scraping components
func (c *KubeAgentMonitor) restartKubeAgent(ctx context.Context, componentType ComponentType) error {

	updateConfigMapErr := c.updateConfigMap(ctx, componentType)
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
func (c *KubeAgentMonitor) updateConfigMap(ctx context.Context, componentType ComponentType) error {

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

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	resp, err := http.Get(baseURL.String())
	if err != nil {
		c.logger.Error("failed to call Restart-API", zap.String("url", baseURL.String()), zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("failed to call get configuration api", zap.Int("statuscode", resp.StatusCode))
		return ErrRestartStatusAPINotOK
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to reas response body", zap.Error(err))
		return err
	}

	// Unmarshal JSON response into ApiResponse struct
	var apiResponse apiResponseForYAML
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		c.logger.Error("failed to unmarshal api response", zap.Error(err))
		return err
	}

	// Verify API Response
	if !apiResponse.Status {
		c.logger.Error("failure status from api response for ingestion rules", zap.Bool("status", apiResponse.Status))
		return ErrInvalidResponse
	}

	var apiYAMLConfig map[string]interface{}
	if len(apiResponse.Config.DaemonSet) == 0 && len(apiResponse.Config.Deployment) == 0 {
		c.logger.Error("failed to get valid response",
			zap.Int("config docker len", len(apiResponse.Config.Docker)),
			zap.Int("config no docker len", len(apiResponse.Config.NoDocker)))
		return ErrInvalidResponse
	}

	apiYAMLConfig = apiResponse.Config.Deployment
	if componentType == DaemonSet {
		apiYAMLConfig = apiResponse.Config.DaemonSet
	}

	yamlData, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		c.logger.Error("failed to marshal api data", zap.Error(err))
		return err
	}

	switch componentType {
	case DaemonSet:
		// Retrieve the existing ConfigMap
		existingDaemonsetConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Get(context.Background(), c.DaemonsetConfigMap, metav1.GetOptions{})
		if err != nil {
			c.logger.Error("Error getting ConfigMap: %v\n" + err.Error())
			return err
		}

		// Modify the content of the ConfigMap
		existingDaemonsetConfigMap.Data["otel-config"] = string(yamlData)

		// Update the ConfigMap
		updatedConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Update(context.Background(), existingDaemonsetConfigMap, metav1.UpdateOptions{})
		if err != nil {
			c.logger.Error("Error updating ConfigMap ", zap.String("error", err.Error()))
			return err
		}

		c.logger.Info("ConfigMap updated successfully ", zap.String("configmap", updatedConfigMap.Name))
	case Deployment:
		// Retrieve the existing ConfigMap
		existingDeploymentConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Get(context.Background(), c.DeploymentConfigMap, metav1.GetOptions{})
		if err != nil {
			c.logger.Error("Error getting ConfigMap ", zap.String("error", err.Error()))
			return err
		}

		// Modify the content of the ConfigMap
		existingDeploymentConfigMap.Data["otel-config"] = string(yamlData)

		// Update the ConfigMap
		updatedDeploymentConfigMap, err := c.Clientset.CoreV1().ConfigMaps(c.AgentNamespace).Update(context.Background(), existingDeploymentConfigMap, metav1.UpdateOptions{})
		if err != nil {
			c.logger.Error("Error updating ConfigMap", zap.String("error", err.Error()))
			return err
		}

		c.logger.Info("ConfigMap updated successfully", zap.String("configmap", updatedDeploymentConfigMap.Name))
	}

	return nil

}
