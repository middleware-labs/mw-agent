package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"

	"strconv"
	"time"

	// "math/rand"

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
	"gopkg.in/yaml.v3"
)

func removeElement(array []string, element string) []string {
	var result []string
	for _, s := range array {
		if s != element {
			result = append(result, s)
		}
	}
	return result
}

func getIndex(slice []string, item string) int {
	for i := range slice {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

func main() {

	go func() {

		for {
			time.Sleep(time.Second * 20)

			rand.Seed(time.Now().UnixNano())

			// Generate a random integer between 0 and 1
			randomInt := rand.Intn(2)

			// Convert the integer to a boolean value
			randomBool := randomInt == 1

			pauseFilelog := randomBool
			fmt.Println("random----", pauseFilelog)

			// reading otel-config.yaml
			configFile, err := os.Open("./configyamls/all/otel-config.yaml")
			if err != nil {
				fmt.Errorf("Could not read config.yaml")
			}
			defer configFile.Close()

			decoder := yaml.NewDecoder(configFile)
			var config OtelStruct
			if err := decoder.Decode(&config); err != nil {
				fmt.Errorf("Could not decode config.yaml")
			}

			// pausing infra file logs
			// removing "filelog" entry from otel-config.yaml log receivers
			if pauseFilelog {
				config.Service.Pipelines.Logs.Receivers = removeElement(config.Service.Pipelines.Logs.Receivers, "filelog")
				fmt.Println("filelogs OFF")
			} else {
				if getIndex(config.Service.Pipelines.Logs.Receivers, "filelog") == -1 {
					config.Service.Pipelines.Logs.Receivers = append(config.Service.Pipelines.Logs.Receivers, "filelog")
				}
				fmt.Println("filelogs ON")
			}

			fmt.Printf("%+v\n", config.Service.Pipelines.Logs.Receivers)

			// writing back otel-config.yaml
			configFile, err = os.Create("./configyamls/all/otel-config.yaml")
			if err != nil {
				panic(err)
			}
			defer configFile.Close()

			encoder := yaml.NewEncoder(configFile)
			if err := encoder.Encode(config); err != nil {
				panic(err)
			}

			fmt.Println("Config updated successfully")

			pid := os.Getpid()
			cmd := exec.Command("kill", "-SIGHUP", fmt.Sprintf("%d", pid))
			if err := cmd.Run(); err != nil {
				fmt.Errorf("Could not load updated otel-config.yaml")
			}
		}

	}()

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
							URIs: []string{configFile},
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
