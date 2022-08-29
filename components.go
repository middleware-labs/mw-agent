package main

import (
	"collector-agent/fluentforwardreceiver"
	"collector-agent/hostmetricsreceiver"
	"log"
	"os"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
)

func Components() (component.Factories, error) {
	var err error
	factories := component.Factories{}
	log.Println("factories.Extensions XXXXXXXX233 setup......")
	log.Println("TARGET ===> ", os.Getenv("TARGET"))
	log.Println("MELT_API_KEY ===> ", os.Getenv("MELT_API_KEY"))
	factories.Extensions, err = component.MakeExtensionFactoryMap(
	// frontend.NewAuthFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Receivers, err = component.MakeReceiverFactoryMap([]component.ReceiverFactory{
		otlpreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Exporters, err = component.MakeExporterFactoryMap([]component.ExporterFactory{
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Processors, err = component.MakeProcessorFactoryMap([]component.ProcessorFactory{
		// frontend.NewProcessorFactory(),
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	return factories, nil
}
