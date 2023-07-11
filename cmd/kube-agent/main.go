package main

import (
	"context"
	"fmt"
	"io/fs"
	"k8sgpt/cmd/analyze"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"os/exec"

	"github.com/prometheus/common/version"
	"github.com/sirupsen/logrus"
	"github.com/tejaskokje-mw/mw-agent/pkg/config"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
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

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// k8sgpt Daily cycle
	go func() {
		// One time execution at the start of the agent
		analyze.GptAnalysis()

		// Running analysis once everyday
		for range time.Tick(time.Second * 10800) {
			analyze.GptAnalysis()
		}
	}()

	_, exists := os.LookupEnv("MW_ISSET_AGENT_INSTALLATION_TIME")
	if exists {
		fmt.Println("env already exists")
		if os.Getenv("MW_ISSET_AGENT_INSTALLATION_TIME") != "true" {
			os.Setenv("MW_ISSET_AGENT_INSTALLATION_TIME", "true")
			os.Setenv("MW_AGENT_INSTALLATION_TIME", strconv.FormatInt(time.Now().UnixMilli(), 10))
		}
	} else {
		fmt.Println("env does not exists")
		cmd := exec.Command("echo MW_ISSET_AGENT_INSTALLATION_TIME=true > /etc/environment")
		stdout, err := cmd.Output()
		if err != nil {
			fmt.Println(err.Error())
		}
		// Print the output
		fmt.Println("1111", string(stdout))
		os.Setenv("MW_ISSET_AGENT_INSTALLATION_TIME", "true")
		os.Setenv("MW_AGENT_INSTALLATION_TIME", strconv.FormatInt(time.Now().UnixMilli(), 10))
	}

	if err := app(logger).Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("could not run application")
	}
}

func Try[T any](item T, err error) T {
	if err != nil {
		log.Fatalf("error %v", err)
	}
	return item
}

func IsSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

// air --build.cmd "go build -o /tmp/api-server /app/*.go" --build.bin "/tmp/api-server $*"
func app(logger *zap.Logger) *cli.App {

	_, hasMwDockerEndpoint := os.LookupEnv("MW_DOCKER_ENDPOINT")
	if !hasMwDockerEndpoint || os.Getenv("MW_DOCKER_ENDPOINT") == "" {
		os.Setenv("MW_DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

	// configure flags
	var apiKey, target, configCheckInterval, apiURLForConfigCheck string
	var enableSyntheticMonitoring bool
	flags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-key",
			EnvVars:     []string{"MW_API_KEY"},
			Usage:       "Middleware API key for your account.",
			Destination: &apiKey,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "target",
			EnvVars:     []string{"MW_TARGET"},
			Usage:       "Middleware target for your account.",
			Destination: &target,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-synthetic-monitoring",
			EnvVars:     []string{"MW_ENABLE_SYNTHETIC_MONITORING"},
			Destination: &enableSyntheticMonitoring,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "config-check-interval",
			EnvVars: []string{"MW_CONFIG_CHECK_INTERVAL"},
			Usage: "Duration string to periodically check for configuration updates." +
				"Setting the value to 0 disables this feature.",
			Destination: &configCheckInterval,
			DefaultText: "60s",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-url-for-config-check",
			EnvVars:     []string{"MW_API_URL_FOR_CONFIG_CHECK"},
			Destination: &apiURLForConfigCheck,
			DefaultText: "https://app.middleware.io",
			Hidden:      true,
		}),

		&cli.StringFlag{
			Name:    "config-file",
			EnvVars: []string{"MW_CONFIG_FILE"},
			Usage:   "Location of the configuration file for this agent.",
			DefaultText: func() string {
				switch runtime.GOOS {
				case "linux":
					return filepath.Join("/etc", "mw-agent", "mw-agent.yaml")
				}

				return ""
			}(),
		},
	}

	return &cli.App{
		Name:  "mw-agent",
		Usage: "Middleware Kubernetes agent",
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "start",
				Usage: "Start Middleware Kubernetes agent",
				Flags: flags,
				Action: func(c *cli.Context) error {

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					cfg := config.NewKubeAgent(
						config.WithKubeAgentApiKey(apiKey),
						config.WithKubeAgentTarget(target),
						config.WithKubeAgentEnableSyntheticMonitoring(
							enableSyntheticMonitoring),
						config.WithKubeAgentConfigCheckInterval(
							configCheckInterval),
						config.WithKubeAgentApiURLForConfigCheck(
							apiURLForConfigCheck),
						config.WithKubeAgentLogger(logger),
					)

					logger.Info("starting host agent with config",
						zap.String("api-key", cfg.ApiKey),
						zap.String("target", cfg.Target),
						zap.String("config-check-interval", cfg.ConfigCheckInterval),
						zap.String("api-url-for-config-check", cfg.ApiURLForConfigCheck))

					yamlPath, err := cfg.GetUpdatedYAMLPath()
					if err != nil {
						logger.Error("error getting config file path", zap.Error(err))
						return err
					}
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
						logger.Error("config provider error", zap.Error(err))
						return err
					}

					factories, err := cfg.GetFactories(ctx)
					if err != nil {
						logger.Error("failed to get factories", zap.Error(err))
						return err
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
						Factories:      factories,
						ConfigProvider: configProvider,
					}
					collector, _ := otelcol.NewCollector(settings)
					if err := collector.Run(context.Background()); err != nil {
						logger.Error("collector server run finished with error", zap.Error(err))
						return err
					}

					return nil
				},
			},
		},
	}
}
