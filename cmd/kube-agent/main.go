package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/k8sgpt-ai/k8sgpt/pkg/kubernetes"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/middleware-labs/mw-agent/pkg/mwinsight"
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
			Name:        "docker-endpoint",
			EnvVars:     []string{"MW_DOCKER_ENDPOINT"},
			Usage:       "Set the endpoint for Docker socket if different from default",
			Destination: &cfg.DockerEndpoint,
			DefaultText: "unix:///var/run/docker.sock",
			Value:       "unix:///var/run/docker.sock",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-url-for-config-check",
			EnvVars:     []string{"MW_API_URL_FOR_CONFIG_CHECK"},
			Destination: &cfg.APIURLForConfigCheck,
			DefaultText: "https://app.middleware.io",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "insight-refresh-duration",
			EnvVars:     []string{"MW_INSIGHT_REFRESH_DURATION"},
			Destination: &cfg.InsightRefreshDuration,
			DefaultText: "24h",
			Value:       "24h",
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

					var wg sync.WaitGroup
					ctx, cancel := context.WithCancel(c.Context)
					defer func() {
						cancel()
						wg.Wait()
					}()

					kubeAgent := agent.NewKubeAgent(cfg,
						agent.WithKubeAgentLogger(logger),
					)

					// Set MW_TARGET & MW_API_KEY so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", cfg.Target)
					os.Setenv("MW_API_KEY", cfg.APIKey)

					// Set MW_DOCKER_ENDPOINT env variable to be used by otel collector
					os.Setenv("MW_DOCKER_ENDPOINT", cfg.DockerEndpoint)

					yamlPath, err := kubeAgent.GetUpdatedYAMLPath()
					if err != nil {
						logger.Error("error getting config file path", zap.Error(err))
						return err
					}

					k8sClient, err := kubernetes.NewClient("", "")
					if err != nil {
						logger.Error("error creating k8s client", zap.Error(err))
						return err
					}

					k8sInsight := mwinsight.NewK8sInsight(
						mwinsight.WithK8sInsightAPIKey(cfg.APIKey),
						mwinsight.WithK8sInsightTarget(cfg.Target),
						mwinsight.WithK8sInsightK8sClient(k8sClient),
						mwinsight.WithK8sInsightBackend(mwinsight.BackendTypeOpenAI),
					)

					logger.Info("starting host agent with config",
						zap.Stringer("config", cfg))

					// start daily insight analysis
					duration, err := time.ParseDuration(cfg.InsightRefreshDuration)
					if err != nil {
						logger.Error("error in parsing duration", zap.Error(err))
						return err
					}

					wg.Add(1)
					go func(ctx context.Context, duration time.Duration, wg *sync.WaitGroup) {
						defer wg.Done()

						// define analysis function that can be reused first time when
						// the agent is run and periodically
						analysisFunc := func() {
							// save current timestamp in the context so that all results of analysis
							// have the same time
							ctx = context.WithValue(ctx, mwinsight.TimeStampCtxKey,
								time.Now())
							analysisChan, err := k8sInsight.Analyze(ctx)
							if err != nil {
								logger.Error("error in mwinsight analysis", zap.Error(err))
								return
							}

							var sendWg sync.WaitGroup
							for result := range analysisChan {
								sendWg.Add(1)
								go func(result []byte) {
									defer sendWg.Done()
									er := k8sInsight.Send(ctx, result)
									if er != nil {
										logger.Error("error sending insight data to mw backend", zap.Error(err))
									}
								}(result)
							}

							sendWg.Wait()
						}

						// run the analysis for the first time after agent
						// start
						analysisFunc()
						select {
						case <-ctx.Done():
							return
						case <-time.Tick(duration):
							// run analysis periodically
							analysisFunc()
						}

					}(ctx, duration, &wg)

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
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}
