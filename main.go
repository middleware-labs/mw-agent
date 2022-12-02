package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	expandconverter "go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
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

	_, hasMwDockerEndpoint := os.LookupEnv("MW_DOCKER_ENDPOINT")
	if !hasMwDockerEndpoint {
		os.Setenv("MW_DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

	_, haveMwLogPaths := os.LookupEnv("MW_LOG_PATHS")
	if !haveMwLogPaths {
		os.Setenv("MW_LOG_PATHS", "")
	}

	collectionType := "all"
	_, err := os.Stat("/var/run/docker.sock")
	if err != nil {
		collectionType = "nodocker"
	}

	value, hasCollectionType := os.LookupEnv("MW_COLLECTION_TYPE")
	if hasCollectionType {
		collectionType = value
	}
	configFile := ""
	configFile = "configyamls/" + collectionType + "/otel-config.yaml"

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
						ResolverSettings: confmap.ResolverSettings{
							Providers: map[string]confmap.Provider{
								"file": fileprovider.New(),
								"yaml": yamlprovider.New(),
								"env":  envprovider.New(),
							},
							Converters: []confmap.Converter{
								expandconverter.New(),
								//overwritepropertiesconverter.New(getSetFlag()),
							},
							URIs: []string{configFile},
						},
					})
					if err != nil {
						log.Fatalf("config provider error thrown %v", err.Error())
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
