package agent

import (
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/fileexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/cumulativetodeltaprocessor"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/deltatorateprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/apachereceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/elasticsearchreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/kafkametricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mysqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/oracledbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redisreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/statsdreceiver"

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
)

// GetFactories get otel factories for HostAgent
func (c *HostAgent) getFactories() (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}
	factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
		// frontend.NewAuthFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	receiverfactories := []receiver.Factory{
		kafkametricsreceiver.NewFactory(),
		otlpreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
		mysqlreceiver.NewFactory(),
		elasticsearchreceiver.NewFactory(),
		redisreceiver.NewFactory(),
		jmxreceiver.NewFactory(),
		apachereceiver.NewFactory(),
		oracledbreceiver.NewFactory(),
		statsdreceiver.NewFactory(),
	}

	// if the host agent is running on ECS EC2, add
	// relevant factories
	if c.InfraPlatform == InfraPlatformECSEC2 || c.InfraPlatform == InfraPlatformECSFargate {
		receiverfactories = append(receiverfactories,
			awsecscontainermetricsreceiver.NewFactory())
	}

	factories.Receivers, err = receiver.MakeFactoryMap(receiverfactories...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = exporter.MakeFactoryMap([]exporter.Factory{
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		kafkaexporter.NewFactory(),
		fileexporter.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = processor.MakeFactoryMap([]processor.Factory{
		// frontend.NewProcessorFactory(),
		metricstransformprocessor.NewFactory(),
		batchprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
		transformprocessor.NewFactory(),
		cumulativetodeltaprocessor.NewFactory(),
		deltatorateprocessor.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}
