package frontend

import (
	"context"
	"go.opentelemetry.io/collector/config/configauth"
	"go.opentelemetry.io/collector/consumer"
	"log"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
)

const typeStr = "auth"

func NewAuthFactory() component.ExtensionFactory {
	return component.NewExtensionFactory(
		typeStr,
		func() config.Extension {
			log.Println("NewAuthFactory config.Extension")
			return &ConfigAuth{
				ExtensionSettings: config.NewExtensionSettings(config.NewComponentID(typeStr)),
			}
		},
		func(ctx context.Context, settings component.ExtensionCreateSettings, extension config.Extension) (component.Extension, error) {
			log.Println("NewAuthFactory Auth")
			xx := &Auth{
				config: extension.(*ConfigAuth),
				logger: settings.Logger,
			}
			return configauth.NewServerAuthenticator(configauth.WithStart(xx.Start), configauth.WithAuthenticate(xx.Authenticate)), nil
		},
	)
}

func NewProcessorFactory() component.ProcessorFactory {
	return component.NewProcessorFactory(
		"frontend",
		func() config.Processor {
			log.Println("NewProcessorFactory config.Extension")

			return &ConfigProcessor{
				ProcessorSettings: config.NewProcessorSettings(config.NewComponentID("frontend")),
			}
		},
		component.WithTracesProcessor(func(ctx context.Context, settings component.ProcessorCreateSettings, cfg config.Processor, traces consumer.Traces) (component.TracesProcessor, error) {

			log.Println("NewProcessorFactory WithTracesProcessor")
			return &MetaProcessor{
				config: cfg.(*ConfigProcessor),
				logger: settings.Logger,
			}, nil
		}),
		component.WithLogsProcessor(func(ctx context.Context, settings component.ProcessorCreateSettings, cfg config.Processor, logs consumer.Logs) (component.LogsProcessor, error) {

			log.Println("NewProcessorFactory WithLogsProcessor")
			return &MetaProcessor{
				config: cfg.(*ConfigProcessor),
				logger: settings.Logger,
			}, nil
		}),
		component.WithMetricsProcessor(func(ctx context.Context, settings component.ProcessorCreateSettings, cfg config.Processor, metrics consumer.Metrics) (component.MetricsProcessor, error) {

			log.Println("NewProcessorFactory WithMetricsProcessor")
			return &MetaProcessor{
				config: cfg.(*ConfigProcessor),
				logger: settings.Logger,
			}, nil
		}),
	)
}
