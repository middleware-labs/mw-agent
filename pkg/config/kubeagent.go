package config

import (
	"context"
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
)

// KubeAgent implements Agent interface for Kubernetes
type KubeAgent struct {
	apiKey string
	target string

	enableSytheticMonitoring bool
	configCheckInterval      string

	apiURLForConfigCheck string

	logger         *zap.Logger
	dockerEndpoint string
}

type KubeOptions func(h *KubeAgent)

func WithKubeAgentApiKey(key string) KubeOptions {
	return func(h *KubeAgent) {
		h.apiKey = key
	}
}

func WithKubeAgentTarget(t string) KubeOptions {
	return func(h *KubeAgent) {
		h.target = t
	}
}

func WithKubeAgentEnableSyntheticMonitoring(e bool) KubeOptions {
	return func(h *KubeAgent) {
		h.enableSytheticMonitoring = e
	}
}

func WithKubeAgentConfigCheckInterval(c string) KubeOptions {
	return func(h *KubeAgent) {
		h.configCheckInterval = c
	}
}

func WithKubeAgentApiURLForConfigCheck(u string) KubeOptions {
	return func(h *KubeAgent) {
		h.apiURLForConfigCheck = u
	}
}

func WithKubeAgentLogger(logger *zap.Logger) KubeOptions {
	return func(h *KubeAgent) {
		h.logger = logger
	}
}

func WithKubeAgentDockerEndpoint(endpoint string) KubeOptions {
	return func(h *KubeAgent) {
		h.dockerEndpoint = endpoint
	}
}

func NewKubeAgent(opts ...KubeOptions) *KubeAgent {
	var cfg KubeAgent
	for _, apply := range opts {
		apply(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger, _ = zap.NewProduction()
	}

	return &cfg
}

func (k *KubeAgent) GetUpdatedYAMLPath() (string, error) {
	yamlPath := "/app/otel-config.yaml"
	dockerSocketPath := strings.Split(k.dockerEndpoint, "//")

	if len(dockerSocketPath) != 2 || !isSocketFn(dockerSocketPath[1]) {
		yamlPath = "/app/otel-config-nodocker.yaml"
	}

	return yamlPath, nil
}

func (k *KubeAgent) GetFactories(ctx context.Context) (otelcol.Factories, error) {
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
