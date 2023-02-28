package main

import (
	"log"
	"os"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/prometheusexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/servicegraphprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func Components() (component.Factories, error) {
	var err error
	factories := component.Factories{}
	log.Println("factories.Extensions XXXXXXXX233 setup......")
	log.Println("TARGET ===> ", os.Getenv("TARGET"))
	log.Println("MW_API_KEY ===> ", os.Getenv("MW_API_KEY"))
	log.Println("Agent Installed At ===> ", os.Getenv("MW_AGENT_INSTALLATION_TIME"))
	factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
	// frontend.NewAuthFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Receivers, err = receiver.MakeFactoryMap([]component.ReceiverFactory{
		otlpreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Exporters, err = exporter.MakeFactoryMap([]component.ExporterFactory{
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
		prometheusexporter.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Processors, err = component.MakeProcessorFactoryMap([]component.ProcessorFactory{
		// frontend.NewProcessorFactory(),
		batchprocessor.NewFactory(),
		servicegraphprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	return factories, nil
}
