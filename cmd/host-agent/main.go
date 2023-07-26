package main

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"checkagent"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/prometheus/common/version"

	"github.com/kardianos/service"
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

type program struct {
	logger *zap.Logger
	args   []string
}

// Service interface for kardianos/service package to run
// on Linux, Windows, MacOS & BSD
func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	return nil
}

func (p *program) run() {
	if err := app(p.logger).Run(p.args); err != nil {
		p.logger.Fatal("could not run application", zap.Error(err))
	}
}
func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	svcConfig := &service.Config{
		Name:        "MiddlewareHostAgent",
		DisplayName: "Middleware Host Agent",
		Description: "Middleware Host Agent for collecting observability signals.",
	}

	prg := &program{
		logger: logger,
		args:   os.Args,
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Fatal("could not create OS service", zap.Error(err))
	}

	err = s.Run()
	if err != nil {
	}
}

func app(logger *zap.Logger) *cli.App {
	var apiKey, target, configCheckInterval, apiURLForConfigCheck string
	var hostTags string
	var dockerEndpoint string
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
			Name:        "docker-endpoint",
			EnvVars:     []string{"MW_DOCKER_ENDPOINT"},
			Usage:       "Set the endpoint for Docker socket if different from default",
			Destination: &dockerEndpoint,
			DefaultText: "unix:///var/run/docker.sock",
			Value:       "unix:///var/run/docker.sock",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-url-for-config-check",
			EnvVars:     []string{"MW_API_URL_FOR_CONFIG_CHECK"},
			Destination: &apiURLForConfigCheck,
			DefaultText: "https://app.middleware.io",
			Value:       "https://app.middleware.io",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "host-tags",
			EnvVars:     []string{"MW_HOST_TAGS"},
			Destination: &hostTags,
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

					hostAgent := agent.NewHostAgent(
						agent.WithHostAgentApiKey(apiKey),
						agent.WithHostAgentTarget(target),
						agent.WithHostAgentEnableSyntheticMonitoring(
							enableSyntheticMonitoring),
						agent.WithHostAgentConfigCheckInterval(
							configCheckInterval),
						agent.WithHostAgentApiURLForConfigCheck(
							apiURLForConfigCheck),
						agent.WithHostAgentLogger(logger),
						agent.WithHostAgentDockerEndpoint(dockerEndpoint),
						agent.WithHostAgentHostTags(hostTags),
					)

					logger.Info("starting host agent with config",
						zap.String("api-key", apiKey),
						zap.String("target", target),
						zap.String("config-check-interval", configCheckInterval),
						zap.Bool("enable-synthetic-monitoring", enableSyntheticMonitoring),
						zap.String("api-url-for-config-check", apiURLForConfigCheck),
						zap.String("docker-endpoint", dockerEndpoint),
						zap.String("host-tags", hostTags))

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					// Listen to the config changes provided by Middleware API
					if configCheckInterval != "0" {
						hostAgent.ListenForConfigChanges(ctx)
					}

					if enableSyntheticMonitoring {
						// TODO checkagent.Start should take context
						go checkagent.Start()
					}

					u, err := url.Parse(target)
					if err != nil {
						return err
					}

					target := u.String()
					if u.Port() == "" {
						target += ":443"
					}

					// Set MW_TARGET & MW_API_KEY so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", target)
					os.Setenv("MW_API_KEY", apiKey)

					// TODO: check if on Windows, socket scheme is different than "unix"
					os.Setenv("MW_DOCKER_ENDPOINT", dockerEndpoint)

					// Setting MW_HOST_TAGS so that envprovider can fill those in the otel config files
					os.Setenv("MW_HOST_TAGS", hostTags)

					// Checking if host agent has valid tags
					if !hostAgent.HasValidTags() {
						return agent.ErrInvalidHostTags
					}

					/*yamlPath, err := hostAgent.GetUpdatedYAMLPath()
					if err != nil {
						logger.Error("error getting config file path", zap.Error(err))
						return err
					}*/

					yamlPath := "./configyamls/all/otel-config.yaml"
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

					factories, err := hostAgent.GetFactories(ctx)
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
