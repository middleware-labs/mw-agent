package main

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"checkagent"

	"github.com/prometheus/common/version"
	"github.com/middleware-labs/mw-agent/pkg/config"

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

	if err := app(logger).Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}

// air --build.cmd "go build -o /tmp/api-server /app/*.go" --build.bin "/tmp/api-server $*"
func app(logger *zap.Logger) *cli.App {

	_, hasMwDockerEndpoint := os.LookupEnv("MW_DOCKER_ENDPOINT")
	if !hasMwDockerEndpoint {
		os.Setenv("MW_DOCKER_ENDPOINT", "unix:///var/run/docker.sock")
	}

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
			Value:       "https://app.middleware.io",
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
		Usage: "Middleware host agent",
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start Middleware host agent",
				Flags:  flags,
				Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config-file")),
				Action: func(c *cli.Context) error {

					cfg := config.NewHostAgent(
						config.WithHostAgentApiKey(apiKey),
						config.WithHostAgentTarget(target),
						config.WithHostAgentEnableSyntheticMonitoring(
							enableSyntheticMonitoring),
						config.WithHostAgentConfigCheckInterval(
							configCheckInterval),
						config.WithHostAgentApiURLForConfigCheck(
							apiURLForConfigCheck),
						config.WithHostAgentLogger(logger),
					)

					logger.Info("starting host agent with config",
						zap.String("api-key", cfg.ApiKey),
						zap.String("target", cfg.Target),
						zap.String("config-check-interval", cfg.ConfigCheckInterval),
						zap.Bool("enable-synthetic-monitoring", cfg.EnableSytheticMonitoring),
						zap.String("api-url-for-config-check", cfg.ApiURLForConfigCheck))

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					// Listen to the config changes provided by Middleware API
					if cfg.ConfigCheckInterval != "0" {
						cfg.ListenForConfigChanges(ctx)
					}

					if cfg.EnableSytheticMonitoring {
						// TODO checkagent.Start should take context
						go checkagent.Start()
					}

					u, err := url.Parse(cfg.Target)
					if err != nil {
						return err
					}

					target := u.String()
					if u.Port() != "" {
						target += ":" + u.Port()
					} else {
						target += ":443"
					}

					// Set MW_TARGET & MW_API_KEY so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", target)
					os.Setenv("MW_API_KEY", cfg.ApiKey)

					yamlPath, err := cfg.GetUpdatedYAMLPath()
					if err != nil {
						logger.Error("error getting config file path", zap.Error(err))
						return err
					}

					// yamlPath := "./configyamls/all/otel-config.yaml"
					logger.Info("yaml path loaded", zap.String("path", yamlPath))

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
							Command:     "mw-otelcontribcol",
							Description: "Middleware OpenTelemetry Collector Contrib",
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
