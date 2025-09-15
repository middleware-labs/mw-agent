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
		altsrc.NewUintFlag(&cli.UintFlag{
			Name:        "agent-internal-metrics-port",
			Usage:       "Port where mw-agent will expose its Prometheus metrics.",
			EnvVars:     []string{"MW_AGENT_INTERNAL_METRICS_PORT"},
			Destination: &cfg.InternalMetricsPort,
			DefaultText: "8888",
			Value:       8888,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "opsai.api-url",
			EnvVars:     []string{"MW_OPSAI_API_URL"},
			Destination: &cfg.OpsAI.ApiURL,
			DefaultText: "wss://app.middleware.io/plsrws/v2",
			Value:       "wss://app.middleware.io/plsrws/v2",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "opsai.unsubscribe-endpoint",
			EnvVars:     []string{"MW_OPSAI_UNSUBSCRIBE_ENDPOINT"},
			Destination: &cfg.OpsAI.UnsubscribeEndpoint,
			DefaultText: "https://app.middleware.io/api/v1/synthetics/unsubscribe",
			Value:       "https://app.middleware.io/api/v1/synthetics/unsubscribe",
			Hidden:      true,
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.opsai-autofix",
			Usage:       "Flag to enable or disable OpsAI Pulsar Connection.",
			EnvVars:     []string{"MW_AGENT_FEATURES_OPSAI_AUTOFIX"},
			Destination: &cfg.AgentFeatures.OpsAIAutoFix,
			DefaultText: "true",
			Value:       true, // synthetic monitoring is disabled by default
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
		Usage: "Middleware Kubernetes OpsAI Agent",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start Middleware Kubernetes OpsAI Agent",
				Flags: flags,
				Action: func(c *cli.Context) error {
					if cfg.SelfProfiling {
						profiler := agent.NewProfiler(logger, cfg.ProfilngServerURL)
						// start profiling
						go profiler.StartProfiling("mw-kube-agent-opsai", cfg.Target, "")
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
					os.Setenv("MW_AGENT_INTERNAL_METRICS_PORT", strconv.Itoa(int(cfg.InternalMetricsPort)))

					// Set MW_DOCKER_ENDPOINT env variable to be used by otel collector
					os.Setenv("MW_DOCKER_ENDPOINT", cfg.DockerEndpoint)

					logger.Info("starting host agent with config",
						zap.Stringer("config", cfg))

					config := worker.Config{
						Mode:                worker.ModeMCP,
						Token:               cfg.APIKey,
						NCAPassword:         cfg.APIKey,
						Hostname:            os.Getenv("MW_KUBE_CLUSTER_NAME"),
						PulsarHost:          cfg.OpsAI.ApiURL,
						Location:            os.Getenv("MW_KUBE_CLUSTER_NAME"),
						UnsubscribeEndpoint: cfg.OpsAI.UnsubscribeEndpoint,
						CaptureEndpoint:     cfg.Target + "/v1/metrics",
					}

					logger.Info("starting opsai worker: ", zap.String("hostname", os.Getenv("MW_KUBE_CLUSTER_NAME")))
					opsaiWorker, err := worker.New(&config)
					if err != nil {
						logger.Error("Failed to create worker")
					}

					func(ctx context.Context) {
						for {
							select {
							case <-ctx.Done():
								fmt.Println("Turning off the opsai agent...")
								return
							default:
								opsaiWorker.Run()
							}
						}
					}(ctx)

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}
