package agent

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/k8sattributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
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
	yaml "gopkg.in/yaml.v2"
)

var (
	ErrParsingConfig = errors.New("error parsing config")
)

// KubeAgent implements Agent interface for Kubernetes
type KubeAgent struct {
	KubeConfig
	logger *zap.Logger
}

// KubeOptions takes in various options for KubeAgent
type KubeOptions func(h *KubeAgent)

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

// GetUpdatedYAMLPath gets the correct otel configuration file.
func (k *KubeAgent) GetUpdatedYAMLPath() (string, error) {
	yamlPath := "/app/otel-config.yaml"
	dockerSocketPath := strings.Split(k.DockerEndpoint, "//")

	if len(dockerSocketPath) != 2 || !isSocketFn(dockerSocketPath[1]) {
		yamlPath = "/app/otel-config-nodocker.yaml"
	}

	return yamlPath, nil
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
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

func (k *KubeAgent) HandleSidecarReceivers(path string) error {
	// Read the YAML file
	data, err := os.ReadFile(path)
	if err != nil {
		k.logger.Error("Error reading YAML file:", zap.Error(err))
		return err
	}

	// Unmarshal YAML into a map[string]interface{}
	var configMap map[string]interface{}
	err = yaml.Unmarshal(data, &configMap)
	if err != nil {
		k.logger.Error("Error unmarshaling YAML:", zap.Error(err))
		return err

	}

	// Check if "pipelines" key exists and is a map
	service, ok := configMap["service"].(map[interface{}]interface{})
	if !ok {
		k.logger.Error("Error: 'service' key not found or is not a map")
		return ErrParsingConfig
	}

	// Check if "pipelines" key exists and is a map
	pipelines, ok := service["pipelines"].(map[interface{}]interface{})
	if !ok {
		k.logger.Error("Error: 'pipelines' key not found or is not a map")
		return ErrParsingConfig
	}

	// Create a new map for filtered pipelines
	filteredPipelines := make(map[interface{}]interface{})

	// Iterate through the pipelines and filter based on "otlp" in "receivers"
	for name, pipeline := range pipelines {
		pipelineMap, ok := pipeline.(map[interface{}]interface{})
		if !ok {
			continue // Skip pipelines that are not maps
		}

		receivers, ok := pipelineMap["receivers"].([]interface{})
		if !ok {
			continue // Skip pipelines without a "receivers" key or if it's not an array
		}

		// Check if sidecar type is present in the "receivers" array
		for _, receiver := range receivers {
			if receiver == "prometheus" {
				filteredPipelines[name] = pipeline
				break
			}
		}
	}

	// Update the config map with the filtered pipelines
	service["pipelines"] = filteredPipelines

	// Marshal the updated map back to YAML
	updatedYAML, err := yaml.Marshal(&configMap)
	if err != nil {
		k.logger.Error("Error marshaling YAML:", zap.Error(err))
		return err
	}

	// Write the updated YAML to a file or print it to the console
	err = os.WriteFile(path, updatedYAML, 0644)
	if err != nil {
		k.logger.Error("Error writing YAML file:", zap.Error(err))
		return err
	}

	return nil
}
