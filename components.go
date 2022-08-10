package main

import (

	// "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"collector-agent/frontend"
	"collector-agent/hostmetricsreceiver"
	"os"

	// "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"log"
	// "github.com/middleware/agents/agent-host/collector-agent/frontend/kafkaexporter"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"

	//"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	//	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	// processors
	// frontend "collector-agent/frontend"
)

func Components() (component.Factories, error) {
	var err error
	factories := component.Factories{}
	log.Println("factories.Extensions XXXXXXXX233 setup......")
	log.Println("TARGET ===> ", os.Getenv("TARGET"))
	log.Println("MELT_API_KEY ===> ", os.Getenv("MELT_API_KEY"))
	log.Println("MELT_API_KEY ===> ", os.Getenv("MELT_API_KEY"))
	factories.Extensions, err = component.MakeExtensionFactoryMap(
		frontend.NewAuthFactory(),
	)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Receivers, err = component.MakeReceiverFactoryMap([]component.ReceiverFactory{
		otlpreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		// k8sclusterreceiver.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Exporters, err = component.MakeExporterFactoryMap([]component.ExporterFactory{
		// kafkaexporter.NewFactory(),
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	factories.Processors, err = component.MakeProcessorFactoryMap([]component.ProcessorFactory{
		frontend.NewProcessorFactory(),
		batchprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		// attributesprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
	}...)
	if err != nil {
		return component.Factories{}, err
	}

	return factories, nil
}
