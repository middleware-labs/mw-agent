package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"checkagent"
	"strconv"
	"time"

	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	expandconverter "go.opentelemetry.io/collector/confmap/converter/expandconverter"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
)

func main() {

	// Listening to Pulsar topics - specific to this host
	if os.Getenv("MW_RUN_SYNTHETIC_TEST_MODULE") != "false" {
		go func() {
			checkagent.Start()
		}()
	}

	os.Setenv("MW_AGENT_INSTALLATION_TIME", strconv.FormatInt(time.Now().UnixMilli(), 10))
	agent_installation_log()

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

	// Listen to the config changes provided by Middleware API
	listenForConfigChanges()

	_, hasMwDockerEndpoint := os.LookupEnv("MW_DOCKER_ENDPOINT")
	if !hasMwDockerEndpoint {
		os.Setenv("MW_DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

	yamlPath := getUpdatedYAMLPath()

	return &cli.App{
		Name:  "api-server",
		Usage: "The API",
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "start",
				Usage: "start API server",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) error {

					configProvider, err := otelcol.NewConfigProvider(otelcol.ConfigProviderSettings{
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
							URIs: []string{yamlPath},
						},
					})
					if err != nil {
						log.Fatalf("config provider error thrown %v", err.Error())
					}
					settings := otelcol.CollectorSettings{
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
					collector, _ := otelcol.NewCollector(settings)
					if err := collector.Run(context.Background()); err != nil {
						return fmt.Errorf("collector server run finished with error: %w", err)
					}

					return nil
				},
			},
		},
	}
}
