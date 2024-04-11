package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/prometheus/common/version"
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
	"go.uber.org/zap/zapcore"
)

func getFlags(cfg *agent.KubeConfig) []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-key",
			EnvVars:     []string{"MW_API_KEY"},
			Usage:       "Middleware API key for your account.",
			Destination: &cfg.APIKey,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "target",
			EnvVars:     []string{"MW_TARGET", "TARGET"},
			Usage:       "Middleware target for your account.",
			Destination: &cfg.Target,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-synthetic-monitoring",
			EnvVars:     []string{"MW_ENABLE_SYNTHETIC_MONITORING"},
			Destination: &cfg.EnableSyntheticMonitoring,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "config-check-interval",
			EnvVars: []string{"MW_CONFIG_CHECK_INTERVAL"},
			Usage: "Duration string to periodically check for configuration updates." +
				"Setting the value to 0 disables this feature.",
			Destination: &cfg.ConfigCheckInterval,
			DefaultText: "60s",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-url-for-config-check",
			EnvVars:     []string{"MW_API_URL_FOR_CONFIG_CHECK"},
			Destination: &cfg.APIURLForConfigCheck,
			DefaultText: "",
			Value:       "",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "docker-endpoint",
			EnvVars:     []string{"MW_DOCKER_ENDPOINT"},
			Usage:       "Set the endpoint for Docker socket if different from default",
			Destination: &cfg.DockerEndpoint,
			DefaultText: "unix:///var/run/docker.sock",
			Value:       "unix:///var/run/docker.sock",
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.infra-monitoring",
			Usage:       "Flag to enable or disable infrastructure monitoring",
			EnvVars:     []string{"MW_AGENT_FEATURES_INFRA_MONITORING"},
			Destination: &cfg.AgentFeatures.InfraMonitoring,
			Value:       true, // infra monitoring is enabled by default
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "mw-agent-self-profiling",
			Usage:       "For Profiling the agent itself",
			EnvVars:     []string{"MW_AGENT_SELF_PROFILING"},
			Destination: &cfg.SelfProfiling,
			Value:       false,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "mw-profiling-server-url",
			Usage:       "Server Address for redirecting profiling data",
			EnvVars:     []string{"MW_PROFILING_SERVER_URL"},
			Destination: &cfg.ProfilngServerURL,
			Value:       "https://profiling.middleware.io",
			DefaultText: "https://profiling.middleware.io",
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
		&cli.StringFlag{
			Name:        "otel-config-file",
			EnvVars:     []string{"MW_OTEL_CONFIG_FILE"},
			Destination: &cfg.OtelConfigFile,
			Usage:       "Location of the OTEL pipelines configuration file for this agent.",
			Value:       filepath.Join("/app", "otel-config-nodocker.yaml"),
			DefaultText: filepath.Join("/app", "otel-config-nodocker.yaml"),
		},
	}
}

func setAgentInstallationTime(logger *zap.Logger) {
	_, exists := os.LookupEnv("MW_ISSET_AGENT_INSTALLATION_TIME")
	if exists {
		logger.Info("MW_ISSET_AGENT_INSTALLATION_TIME env variable exists")
		if os.Getenv("MW_ISSET_AGENT_INSTALLATION_TIME") != "true" {
			os.Setenv("MW_ISSET_AGENT_INSTALLATION_TIME", "true")
			os.Setenv("MW_AGENT_INSTALLATION_TIME",
				strconv.FormatInt(time.Now().UnixMilli(), 10))
		}
	} else {
		logger.Info("MW_ISSET_AGENT_INSTALLATION_TIME env variable does not exists")
		os.Setenv("MW_ISSET_AGENT_INSTALLATION_TIME", "true")
		os.Setenv("MW_AGENT_INSTALLATION_TIME",
			strconv.FormatInt(time.Now().UnixMilli(), 10))
	}
}

func main() {
	var cfg agent.KubeConfig
	flags := getFlags(&cfg)

	zapEncoderCfg := zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig = zapEncoderCfg
	logger, _ := zapCfg.Build()
	defer func() {
		_ = logger.Sync()
	}()

	setAgentInstallationTime(logger)

	app := &cli.App{
		Name:  "mw-agent",
		Usage: "Middleware Kubernetes agent",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start Middleware Kubernetes agent",
				Flags: flags,
				Action: func(c *cli.Context) error {
					if cfg.SelfProfiling {
						profiler := agent.NewProfiler(logger, cfg.ProfilngServerURL)
						// start profiling
						go profiler.StartProfiling("mw-kube-agent", cfg.Target, "")
					}

					ctx, cancel := context.WithCancel(c.Context)
					defer func() {
						cancel()
					}()

					kubeAgent := agent.NewKubeAgent(cfg,
						agent.WithKubeAgentLogger(logger),
					)

					// Set the infra platform to kubernetes for now since we don't need to differentiate between
					// vanilla kubernetes and managed kubernetes.
					cfg.InfraPlatform = agent.InfraPlatformKubernetes
					// Set MW_TARGET & MW_API_KEY so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", cfg.Target)
					os.Setenv("MW_API_KEY", cfg.APIKey)

					// Set MW_DOCKER_ENDPOINT env variable to be used by otel collector
					os.Setenv("MW_DOCKER_ENDPOINT", cfg.DockerEndpoint)

					logger.Info("starting host agent with config",
						zap.Stringer("config", cfg))

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
							URIs: []string{cfg.OtelConfigFile},
						},
					})
					if err != nil {
						logger.Error("config provider error", zap.Error(err))
						return err
					}

					factories, err := kubeAgent.GetFactories(ctx)
					if err != nil {
						logger.Error("failed to get factories", zap.Error(err))
						return err
					}

					settings := otelcol.CollectorSettings{
						DisableGracefulShutdown: true,
						LoggingOptions:          []zap.Option{},
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
			{
				Name:  "update",
				Usage: "Watch for configuration updates and restart the agent when a change is detected",
				Flags: flags,
				Action: func(c *cli.Context) error {

					if cfg.APIURLForConfigCheck == "" {
						var err error
						cfg.APIURLForConfigCheck, err = agent.GetAPIURLForConfigCheck(cfg.Target)
						// could not derive api url for config check from target
						if err != nil {
							logger.Info("could not derive api url for config check from target",
								zap.String("target", cfg.Target))
							return err
						}

						logger.Info("derived api url for config check",
							zap.String("api-url-for-config-check", cfg.APIURLForConfigCheck))
					}

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					kubeAgentMonitor := agent.NewKubeAgentMonitor(cfg,
						agent.WithKubeAgentMonitorClusterName(os.Getenv("MW_KUBE_CLUSTER_NAME")),
						agent.WithKubeAgentMonitorAgentNamespace("mw-agent-ns"),
						agent.WithKubeAgentMonitorDaemonset("mw-kube-agent"),
						agent.WithKubeAgentMonitorDeployment("mw-kube-agent"),
						agent.WithKubeAgentMonitorDaemonsetConfigMap("mw-daemonset-otel-config"),
						agent.WithKubeAgentMonitorDeploymentConfigMap("mw-deployment-otel-config"),
					)

					err := kubeAgentMonitor.SetClientSet()
					if err != nil {
						logger.Error("collector server run finished with error", zap.Error(err))
						return err
					}
					if cfg.ConfigCheckInterval != "0" {
						err = kubeAgentMonitor.ListenForKubeOtelConfigChanges(ctx)
						if err != nil {
							logger.Info("error for listening for config changes", zap.Error(err))
							return err
						}
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}
