package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	expandconverter "go.opentelemetry.io/collector/config/mapconverter/expandmapconverter"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"

	// "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/kafkaexporter"
	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/service"
	"go.uber.org/zap"
)

func main() {
	if err := app().Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("could not run application")
	}
}

func Try[T any](item T, err error) T {
	if err != nil {
		log.Fatalf("error %v", err)
	}
	return item
}

// air --build.cmd "go build -o /tmp/api-server /app/*.go" --build.bin "/tmp/api-server $*"
func app() *cli.App {
	return &cli.App{
		Name:  "api-server",
		Usage: "The API",
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "start",
				Usage: "start API server",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) error {

					configProvider, err := service.NewConfigProvider(service.ConfigProviderSettings{
						Locations: []string{"otel-config.yaml"},
						MapProviders: map[string]confmap.Provider{
							"file": fileprovider.New(),
							"yaml": yamlprovider.New(),
							"env":  envprovider.New(),
						},
						MapConverters: []confmap.Converter{
							expandconverter.New(),
							//overwritepropertiesconverter.New(getSetFlag()),
						},
					})
					if err != nil {
						log.Fatalf("config provider error thrown ", err.Error())
					}
					p := service.CollectorSettings{
						DisableGracefulShutdown: true,
						LoggingOptions:          []zap.Option{
							// zap.Development(),
							// zap.IncreaseLevel(zap.DebugLevel),
						},
						BuildInfo: component.BuildInfo{
							Command:     "otelcontribcol",
							Description: "OpenTelemetry Collector Contrib",
							Version:     version.Version,
						},
						Factories:      Try(Components()),
						ConfigProvider: configProvider,
					}
					cmdc, _ := service.New(p)
					if err := cmdc.Run(context.TODO()); err != nil {
						return fmt.Errorf("collector server run finished with error: %w", err)
					}

					return nil
				},
			},
		},
	}
}
