package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/middleware-labs/synthetics-agent/pkg/worker"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/kardianos/service"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var agentVersion = "0.0.1"

type program struct {
	logger    *zap.Logger
	hostAgent *agent.HostAgent
	programWG *sync.WaitGroup
	// stop the telemetry collection if errCh receives an error
	// resume when errCh receives nil
	errCh  chan error
	stopCh chan struct{}
	args   []string
}

// Service interface for kardianos/service package to run
// on Linux, Windows, MacOS & BSD
func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.

	p.logger.Info("starting service", zap.Stringer("name", s))

	p.programWG.Add(1)
	go p.run()
	p.programWG.Add(1)
	go func() {
		p.hostAgent.ReportServices(p.errCh, p.stopCh)
		p.programWG.Done()
	}()

	// Start any goroutines that can control collection
	if p.hostAgent.FetchAccountOtelConfig {
		// Listen to the config changes provided by Middleware API
		p.programWG.Add(1)
		go func() {
			p.hostAgent.ListenForConfigChanges(p.errCh, p.stopCh)
			p.programWG.Done()
		}()
	} else {
		p.errCh <- nil
	}

	return nil
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	p.logger.Info("stopping service", zap.Stringer("name", s))

	// stop collection
	close(p.stopCh)
	close(p.errCh)
	p.programWG.Wait()
	return nil
}

func (p *program) run() {
	defer p.programWG.Done()

	for err := range p.errCh {
		// if invalid config is received from the backend, then keep collector
		// in its current state. If it is stopped, keep it stopped until we receive
		// a valid config. If it is already running, don't restart it.
		if errors.Is(err, agent.ErrInvalidConfig) {
			p.logger.Error("invalid config. keeping collector in its current state",
				zap.Error(err))
			continue
		}

		if errors.Is(err, agent.ErrReportApiFailure) {
			p.logger.Error("discovery reporting failed; telemetry collection continuing", zap.Error(err))
			continue
		}

		if err != nil {
			// stop collection only if it's running
			p.hostAgent.StopCollector(err)

			// if err is not agent.ErrRestartAgent, then keep collector stopped.
			// if err is agent.ErrRestartAgent, then resume collection.
			if !errors.Is(err, agent.ErrRestartAgent) {
				continue
			}
			p.logger.Info("restarting collector", zap.Error(err))
		}
		// start collection only if it's not running
		if err := p.hostAgent.StartCollector(); err != nil {
			p.logger.Error("failed to start collector",
				zap.Error(err))
		}
	}
}

func getFlags(execPath string, cfg *agent.HostConfig) []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-key",
			EnvVars:     []string{"MW_API_KEY"},
			Usage:       "Middleware API key for your account.",
			Destination: &cfg.APIKey,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "target",
			EnvVars:     []string{"MW_TARGET"},
			Usage:       "Middleware target for your account.",
			Destination: &cfg.Target,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "config-check-interval",
			EnvVars: []string{"MW_CONFIG_CHECK_INTERVAL"},
			Usage: "Duration string to periodically check for configuration updates. " +
				"Setting the value to 0 disables this feature.",
			Destination: &cfg.ConfigCheckInterval,
			DefaultText: "60s",
			Value:       "60s",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "fetch-account-otel-config",
			EnvVars:     []string{"MW_FETCH_ACCOUNT_OTEL_CONFIG"},
			Usage:       "Get the otel-config from Middleware backend.",
			Destination: &cfg.FetchAccountOtelConfig,
			DefaultText: "true",
			Value:       true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "docker-endpoint",
			EnvVars:     []string{"MW_DOCKER_ENDPOINT"},
			Usage:       "Set the endpoint for Docker socket if different from default.",
			Destination: &cfg.DockerEndpoint,
			DefaultText: "unix:///var/run/docker.sock",
			Value:       "unix:///var/run/docker.sock",
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
			Name:        "synthetic-monitoring.api-url",
			EnvVars:     []string{"MW_SYNTHETIC_MONITORING_API_URL", "MW_API_URL_FOR_SYNTHETIC_MONITORING"},
			Destination: &cfg.SyntheticMonitoring.ApiURL,
			DefaultText: "wss://app.middleware.io/plsrws/v2",
			Value:       "wss://app.middleware.io/plsrws/v2",
			Hidden:      true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "host-tags",
			Usage:       "Tags for this host.",
			EnvVars:     []string{"MW_HOST_TAGS"},
			Destination: &cfg.HostTags,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "logfile",
			Usage:       "Log file to store Middleware agent logs.",
			EnvVars:     []string{"MW_LOGFILE"},
			Destination: &cfg.Logfile,
			DefaultText: "",
			Value:       "",
		}),
		altsrc.NewIntFlag(&cli.IntFlag{
			Name:        "logfile-size",
			Usage:       "Log file size to store Middleware agent logs. This flag only applifes if logfile flag is specified.",
			EnvVars:     []string{"MW_LOGFILE_SIZE"},
			Destination: &cfg.LogfileSize,
			DefaultText: "1",
			Value:       1, // default value is 1MB
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "logging-level",
			Usage:       "Logging level for Middleware agent logs. Valid values: debug, info, warn, error, fatal, panic. Default: info.",
			EnvVars:     []string{"MW_LOGGING_LEVEL"},
			Destination: &cfg.LoggingLevel,
			DefaultText: "info",
			Value:       "info",
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
			Usage:       "Flag to enable or disable infrastructure monitoring.",
			EnvVars:     []string{"MW_AGENT_FEATURES_METRIC_COLLECTION"},
			Destination: &cfg.AgentFeatures.MetricCollection,
			Aliases:     []string{"infra-monitoring"},
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
			Name:        "agent-features.synthetic-monitoring",
			Usage:       "Flag to enable or disable synthetic monitoring.",
			EnvVars:     []string{"MW_AGENT_FEATURES_SYNTHETIC_MONITORING"},
			Destination: &cfg.AgentFeatures.SyntheticMonitoring,
			DefaultText: "false",
			Value:       false, // synthetic monitoring is disabled by default
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-self-profiling",
			Usage:       "For Profiling MW Agent itself.",
			EnvVars:     []string{"MW_AGENT_SELF_PROFILING"},
			Destination: &cfg.SelfProfiling,
			DefaultText: "false",
			Value:       false,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "profiling-server-url",
			Usage:       "Server Address for redirecting profiling data.",
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
			Name:        "service-report-interval",
			EnvVars:     []string{"MW_SERVICE_REPORT_INTERVAL"},
			Usage:       "Interval to report service discovery status.",
			Destination: &cfg.ServiceReportInterval,
			DefaultText: "5m",
			Value:       "5m",
		}),

		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:        "agent-features.service-reporting",
			Usage:       "Enable or disable service discovery reporting.",
			EnvVars:     []string{"MW_REPORT_SERVICES"},
			Destination: &cfg.AgentFeatures.ServiceReporting,
			DefaultText: "true",
			Value:       true,
		}),

		&cli.StringFlag{
			Name:    "config-file",
			EnvVars: []string{"MW_CONFIG_FILE"},
			Usage:   "Location of the configuration file for this agent.",
			DefaultText: func() string {
				switch runtime.GOOS {
				case "linux":
					return filepath.Join("/etc", "mw-agent", "agent-config.yaml")
				case "darwin":
					return filepath.Join("/etc", "mw-agent", "agent-config.yaml")
				case "windows":
					return filepath.Join(filepath.Dir(execPath), "agent-config.yaml")
				}

				return ""
			}(),
		},
		&cli.StringFlag{
			Name:        "otel-config-file",
			EnvVars:     []string{"MW_OTEL_CONFIG_FILE"},
			Destination: &cfg.OtelConfigFile,
			Usage:       "Location of the OTEL pipelines configuration file for this agent.",
			Value: func() string {
				switch runtime.GOOS {
				case "linux":
					return filepath.Join("/etc", "mw-agent", "otel-config.yaml")
				case "darwin":
					return filepath.Join("/etc", "mw-agent", "otel-config.yaml")
				case "windows":
					return filepath.Join(filepath.Dir(execPath), "otel-config.yaml")
				}

				return ""
			}(),
			DefaultText: func() string {
				switch runtime.GOOS {
				case "linux":
					return filepath.Join("/etc", "mw-agent", "otel-config.yaml")
				case "darwin":
					return filepath.Join("/etc", "mw-agent", "otel-config.yaml")
				case "windows":
					return filepath.Join(filepath.Dir(execPath), "otel-config.yaml")
				}

				return ""
			}(),
		},
	}
}

func detectInfraPlatform() agent.InfraPlatform {
	awsEnv := os.Getenv("AWS_EXECUTION_ENV")
	if awsEnv == "AWS_ECS_EC2" {
		return agent.InfraPlatformECSEC2
	} else if awsEnv == "AWS_ECS_FARGATE" {
		return agent.InfraPlatformECSFargate
	}

	cycleIOEnv := os.Getenv("CYCLE_INSTANCE_ID")
	if cycleIOEnv != "" {
		return agent.InfraPlatformCycleIO
	}

	// Check if running on EC2 (but not ECS)
	if agent.IsEC2Instance() {
		return agent.InfraPlatformEC2
	}

	return agent.InfraPlatformInstance
}

func main() {
	zapEncoderCfg := zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,

		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
	}

	zapCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapEncoderCfg),
		zapcore.AddSync(os.Stderr),
		zap.InfoLevel,
	)
	logger := zap.New(zapCore, zap.AddCaller())

	execPath, err := os.Executable()
	if err != nil {
		logger.Error("error getting executable path", zap.Error(err))
		return
	}

	var cfg agent.HostConfig
	flags := getFlags(execPath, &cfg)

	app := &cli.App{
		Name:  "mw-agent",
		Usage: "Middleware host agent",
		Commands: []*cli.Command{
			{
				Name:   "start",
				Usage:  "Start Middleware host agent",
				Flags:  flags,
				Before: altsrc.InitInputSourceWithContext(flags, altsrc.NewYamlSourceFromFlagFunc("config-file")),
				Action: func(c *cli.Context) error {
					loggingLevel, err := zap.ParseAtomicLevel(cfg.LoggingLevel)
					if err != nil {
						logger.Error("error getting executable path", zap.Error(err))
						return err
					}

					var w zapcore.WriteSyncer
					if cfg.Logfile != "" {
						logger.Info("redirecting logs to logfile", zap.String("logfile", cfg.Logfile))
						// logfile specified. Update logger to write logs to the
						// give logfile
						w = zapcore.AddSync(&lumberjack.Logger{
							Filename:   cfg.Logfile,
							MaxSize:    cfg.LogfileSize, // megabytes
							MaxBackups: 1,
							MaxAge:     7, // days
						})

					} else {
						w = zapcore.AddSync(os.Stderr)

					}

					zapCore = zapcore.NewCore(
						zapcore.NewJSONEncoder(zapEncoderCfg),
						w,
						loggingLevel.Level(),
					)

					logger := zap.New(zapCore, zap.AddCaller())

					defer func() {
						_ = logger.Sync()
					}()

					if cfg.SelfProfiling {
						profiler := agent.NewProfiler(logger, cfg.ProfilngServerURL)
						// start profiling
						go profiler.StartProfiling("mw-host-agent", cfg.Target, cfg.HostTags)
					}

					infraPlatform := detectInfraPlatform()

					var hostname string

					// Get hostname based on infrastructure platform
					hostname = agent.GetHostnameForPlatform(infraPlatform)

					logger.Info("starting host agent",
						zap.String("agent location", execPath),
						zap.String("hostname", hostname),
						zap.String("OS", runtime.GOOS),
						zap.String("arch", runtime.GOARCH))

					logger.Info("host agent config", zap.Stringer("config", cfg),
						zap.String("version", agentVersion),
						zap.Stringer("infra-platform", infraPlatform))

					if cfg.APIURLForConfigCheck == "" {
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

					if cfg.SyntheticMonitoring.ApiURL == "" {
						cfg.SyntheticMonitoring.ApiURL, err = agent.GetAPIURLForSyntheticMonitoring(cfg.Target)
						// could not derive api url for synthetic monitoring from target
						if err != nil {
							logger.Info("could not derive api url for synthetic monitoring from target",
								zap.String("target", cfg.Target))
							return err
						}

						logger.Info("derived api url for synthetic monitoring",
							zap.String("api-url-for-synthetic-monitoring", cfg.SyntheticMonitoring.ApiURL))
					}

					u, err := url.Parse(cfg.Target)
					if err != nil {
						return err
					}

					target := u.String()
					if u.Port() == "" {
						target += ":443"
					}

					// Set environment variables so that envprovider can fill those in the otel config files
					os.Setenv("MW_TARGET", target)
					os.Setenv("MW_API_KEY", cfg.APIKey)
					os.Setenv("MW_AGENT_GRPC_PORT", cfg.GRPCPort)
					os.Setenv("MW_AGENT_HTTP_PORT", cfg.HTTPPort)
					os.Setenv("MW_AGENT_FLUENT_PORT", cfg.FluentPort)
					os.Setenv("MW_AGENT_INTERNAL_METRICS_PORT", strconv.Itoa(int(cfg.InternalMetricsPort)))

					if cfg.EnableDataDogReceiver {
						os.Setenv("MW_ENABLE_DATADOG_RECEIVER", "true")
					}

					// TODO: check if on Windows, socket scheme is different than "unix"
					os.Setenv("MW_DOCKER_ENDPOINT", cfg.DockerEndpoint)

					// Setting MW_HOST_TAGS so that envprovider can fill those in the otel config files
					os.Setenv("MW_HOST_TAGS", cfg.HostTags)
					// Checking if host agent has valid tags
					if err := agent.HasValidTags(cfg.HostTags); err != nil {
						logger.Info("host agent has invalid tags", zap.Error(err))
						return err
					}
					// create hostAgent

					hostAgent, err := agent.NewHostAgent(
						cfg, zapCore,
						agent.WithHostAgentVersion(agentVersion),
						agent.WithHostAgentInfraPlatform(infraPlatform),
					)

					if err != nil {
						return err
					}

					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					if cfg.AgentFeatures.SyntheticMonitoring {
						config := worker.Config{
							Mode:                worker.ModeAgent,
							Token:               cfg.APIKey,
							Hostname:            hostname,
							PulsarHost:          cfg.SyntheticMonitoring.ApiURL,
							Location:            hostname,
							UnsubscribeEndpoint: cfg.SyntheticMonitoring.UnsubscribeEndpoint,
							CaptureEndpoint:     cfg.Target + "/v1/metrics",
						}

						logger.Info("starting synthetic worker: ", zap.String("hostname", hostname))
						syntheticWorker, err := worker.New(&config)
						if err != nil {
							logger.Error("Failed to create worker")
						}

						go func(ctx context.Context) {
							for {
								select {
								case <-ctx.Done():
									fmt.Println("Turning off the synthetic monitoring...")
									return
								default:
									syntheticWorker.Run()
								}
							}
						}(ctx)

					}

					svcConfig := &service.Config{
						Name:        "mw-agent",
						DisplayName: "Middleware Host Agent",
						Description: "Middleware Host Agent for collecting observability signals.",
					}

					programWG := &sync.WaitGroup{}
					// errCh is used to control whether the agent should collect telemetry data or not.
					// if any of the module returns error, the agent should not collect telemetry data.
					// For example, if the agent is not able to connect to the target,
					// it should not collect telemetry data until it can connect to the target again.
					errCh := make(chan error)

					// stopCh is used to stop the go routine that can send errors to errCh
					stopCh := make(chan struct{})

					prg := &program{
						logger:    logger,
						hostAgent: hostAgent,
						programWG: programWG,
						errCh:     errCh,
						stopCh:    stopCh,
						args:      os.Args,
					}

					s, err := service.New(prg, svcConfig)
					if err != nil {
						logger.Fatal("could not create OS service", zap.Error(err))
					}

					// Run the service
					err = s.Run()
					if err != nil {
						logger.Error("error after running the service", zap.Error(err))
					}
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Returns the current agent version",
				Action: func(c *cli.Context) error {
					fmt.Println("Middleware Agent Version", agentVersion)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}
