package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/middleware-labs/synthetics-agent/pkg/worker"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
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
		/* infra monitoring flag is deprecated. See log-collection flag */
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.infra-monitoring",
			Usage:       "Flag to enable or disable metric collection",
			EnvVars:     []string{"MW_AGENT_FEATURES_INFRA_MONITORING"},
			Destination: &cfg.AgentFeatures.MetricCollection,
			DefaultText: "true",
			Value:       true, // infra monitoring is enabled by default
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.metric-collection",
			Usage:       "Flag to enable or disable metric collection",
			EnvVars:     []string{"MW_AGENT_FEATURES_METRIC_COLLECTION"},
			Destination: &cfg.AgentFeatures.MetricCollection,
			DefaultText: "true",
			Value:       true, // infra monitoring is enabled by default
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.log-collection",
			Usage:       "Flag to enable or disable log collection.",
			EnvVars:     []string{"MW_AGENT_FEATURES_LOG_COLLECTION"},
			Destination: &cfg.AgentFeatures.LogCollection,
			DefaultText: "true",
			Value:       true, // log collection is enabled by default
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-self-profiling",
			Usage:       "For Profiling the agent itself",
			EnvVars:     []string{"MW_AGENT_SELF_PROFILING"},
			Destination: &cfg.SelfProfiling,
			Value:       false,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "profiling-server-url",
			Usage:       "Server Address for redirecting profiling data",
			EnvVars:     []string{"MW_PROFILING_SERVER_URL"},
			Destination: &cfg.ProfilngServerURL,
			Value:       "https://profiling.middleware.io",
			DefaultText: "https://profiling.middleware.io",
		}),

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "grpc-port",
			Usage:       "gRPC receiver will listen to this port.",
			EnvVars:     []string{"MW_AGENT_GRPC_PORT"},
			Destination: &cfg.GRPCPort,
			DefaultText: "9319",
			Value:       "9319",
		}),

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "http-port",
			Usage:       "HTPT receiver will listen to this port.",
			EnvVars:     []string{"MW_AGENT_HTTP_PORT"},
			Destination: &cfg.HTTPPort,
			DefaultText: "9320",
			Value:       "9320",
		}),

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "fluent-port",
			Usage:       "Fluent receiver will listen to this port.",
			EnvVars:     []string{"MW_AGENT_FLUENT_PORT"},
			Destination: &cfg.FluentPort,
			DefaultText: "8006",
			Value:       "8006",
		}),

		altsrc.NewUintFlag(&cli.UintFlag{
			Name:        "agent-internal-metrics-port",
			Usage:       "Port where mw-agent will expose its Prometheus metrics.",
			EnvVars:     []string{"MW_AGENT_INTERNAL_METRICS_PORT"},
			Destination: &cfg.InternalMetricsPort,
			DefaultText: "8888",
			Value:       8888,
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "enable-datadog-receiver",
			Usage:       "Enable datadog receiver in agent",
			EnvVars:     []string{"MW_ENABLE_DATADOG_RECEIVER"},
			Destination: &cfg.EnableDataDogReceiver,
			DefaultText: "false",
			Value:       false,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "synthetic-monitoring.api-url",
			EnvVars:     []string{"MW_SYNTHETIC_MONITORING_API_URL", "MW_API_URL_FOR_SYNTHETIC_MONITORING"},
			Destination: &cfg.SyntheticMonitoring.ApiURL,
			DefaultText: "wss://app.middleware.io/plsrws/v2",
			Value:       "wss://app.middleware.io/plsrws/v2",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "synthetic-monitoring.unsubscribe-endpoint",
			EnvVars:     []string{"MW_SYNTHETIC_MONITORING_UNSUBSCRIBE_ENDPOINT"},
			Destination: &cfg.SyntheticMonitoring.UnsubscribeEndpoint,
			DefaultText: "https://app.middleware.io/api/v1/synthetics/unsubscribe",
			Value:       "https://app.middleware.io/api/v1/synthetics/unsubscribe",
			Hidden:      true,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.synthetic-monitoring",
			Usage:       "Flag to enable or disable synthetic monitoring.",
			EnvVars:     []string{"MW_AGENT_FEATURES_SYNTHETIC_MONITORING"},
			Destination: &cfg.AgentFeatures.SyntheticMonitoring,
			DefaultText: "false",
			Value:       false, // synthetic monitoring is disabled by default
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

					// Set the infra platform to kubernetes for now since we don't need to differentiate between
					// vanilla kubernetes and managed kubernetes.
					cfg.InfraPlatform = agent.InfraPlatformKubernetes

					// Set environment variables so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", cfg.Target)
					os.Setenv("MW_API_KEY", cfg.APIKey)
					os.Setenv("MW_AGENT_GRPC_PORT", cfg.GRPCPort)
					os.Setenv("MW_AGENT_HTTP_PORT", cfg.HTTPPort)
					os.Setenv("MW_AGENT_FLUENT_PORT", cfg.FluentPort)
					os.Setenv("MW_AGENT_INTERNAL_METRICS_PORT", strconv.Itoa(int(cfg.InternalMetricsPort)))

					// Set MW_DOCKER_ENDPOINT env variable to be used by otel collector
					os.Setenv("MW_DOCKER_ENDPOINT", cfg.DockerEndpoint)

					logger.Info("starting host agent with config",
						zap.Stringer("config", cfg))

					if cfg.AgentFeatures.SyntheticMonitoring {
						config := worker.Config{
							Mode:                worker.ModeAgent,
							Token:               cfg.APIKey,
							NCAPassword:         cfg.APIKey,
							Hostname:            os.Getenv("MW_KUBE_CLUSTER_NAME"),
							PulsarHost:          cfg.SyntheticMonitoring.ApiURL,
							Location:            os.Getenv("MW_KUBE_CLUSTER_NAME"),
							UnsubscribeEndpoint: cfg.SyntheticMonitoring.UnsubscribeEndpoint,
							CaptureEndpoint:     cfg.Target + "/v1/metrics",
						}

						logger.Info("starting synthetics worker: ", zap.String("hostname", os.Getenv("MW_KUBE_CLUSTER_NAME")))
						synWorker, err := worker.New(&config)
						if err != nil {
							logger.Error("Failed to create worker")
						}

						go func(ctx context.Context) {
							for {
								select {
								case <-ctx.Done():
									fmt.Println("Turning off the synthetics agent...")
									return
								default:
									synWorker.Run()
								}
							}
						}(ctx)

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
