package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	configupdater "github.com/middleware-labs/mw-agent/pkg/configupdater"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var agentVersion = "0.0.1"

func getFlags(cfg *configupdater.BaseConfig) []cli.Flag {
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

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:    "config-check-interval",
			EnvVars: []string{"MW_CONFIG_CHECK_INTERVAL"},
			Usage: "Duration string to periodically check for configuration updates." +
				"Setting the value to 0 disables this feature.",
			Destination: &cfg.ConfigCheckInterval,
			DefaultText: "60s",
			Value:       "60s",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-url-for-config-check",
			EnvVars:     []string{"MW_API_URL_FOR_CONFIG_CHECK"},
			Destination: &cfg.APIURLForConfigCheck,
			DefaultText: "",
			Value:       "",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "cluster-name",
			EnvVars:     []string{"MW_KUBE_CLUSTER_NAME"},
			Destination: &cfg.ClusterName,
			DefaultText: "",
			Value:       "",
			Required:    true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "namespace-name",
			EnvVars:     []string{"MW_NAMESPACE_NAME"},
			Destination: &cfg.AgentNamespaceName,
			DefaultText: "mw-agent-ns",
			Value:       "mw-agent-ns",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "daemonset-name",
			EnvVars:     []string{"MW_DAEMONSET_NAME"},
			Destination: &cfg.DaemonsetName,
			DefaultText: "mw-kube-agent",
			Value:       "mw-kube-agent",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "daemonset-configmap-name",
			EnvVars:     []string{"MW_DAEMONSET_CONFIGMAP_NAME"},
			Destination: &cfg.DaemonsetConfigMapName,
			DefaultText: "mw-daemonset-otel-config",
			Value:       "mw-daemonset-otel-config",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "deployment-name",
			EnvVars:     []string{"MW_DEPLOYMENT_NAME"},
			Destination: &cfg.DeploymentName,
			DefaultText: "mw-kube-agent",
			Value:       "mw-kube-agent",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "deployment-configmap-name",
			EnvVars:     []string{"MW_DEPLOYMENT_CONFIGMAP_NAME"},
			Destination: &cfg.DeploymentConfigMapName,
			DefaultText: "mw-deployment-otel-config",
			Value:       "mw-deployment-otel-config",
		}),
	}
}

func main() {
	var cfg configupdater.BaseConfig
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
		Name:  "mw-kube-agent-config-updater",
		Usage: "Middleware Kubernetes Agent Configuration Updater",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Watch for configuration updates and restart the agent when a change is detected",
				Flags: flags,
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					if cfg.APIURLForConfigCheck == "" {
						var err error
						cfg.APIURLForConfigCheck, err = configupdater.GetAPIURLForConfigCheck(cfg.Target)
						// could not derive api url for config check from target
						if err != nil {
							logger.Info("could not derive api url for config check from target",
								zap.String("target", cfg.Target))
							return err
						}

						logger.Info("derived api url for config check",
							zap.String("api-url-for-config-check", cfg.APIURLForConfigCheck))
					}

					logger.Info("creating kube agent config updater", zap.Any("config", cfg))
					config, err := rest.InClusterConfig()
					if err != nil {
						return err
					}

					clientset, err := kubernetes.NewForConfig(config)
					if err != nil {
						return err
					}

					kubeAgentUpdater, err := configupdater.NewKubeAgent(cfg, agentVersion,
						clientset, logger)
					if err != nil {
						logger.Fatal("failed to create kube agent config", zap.Error(err))
						return err
					}

					var wg sync.WaitGroup
					// errCh is used to control whether the agent should collect telemetry data or not.
					// if any of the module returns error, the agent should not collect telemetry data.
					// For example, if the agent is not able to connect to the target,
					// it should not collect telemetry data until it can connect to the target again.
					errCh := make(chan error)

					// stopCh is used to stop the go routine that can send errors to errCh
					stopCh := make(chan struct{})

					wg.Add(1)
					go func() {
						defer wg.Done()
						err := kubeAgentUpdater.ListenForConfigChanges(ctx, errCh, stopCh)
						if err != nil {
							logger.Info("error for listening for config changes", zap.Error(err))
						}
					}()

					for err := range errCh {
						if err != nil {
							logger.Info("error for listening for config changes", zap.Error(err))
						}
					}

					close(stopCh)
					wg.Wait()

					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Returns the current agent version",
				Action: func(c *cli.Context) error {
					fmt.Println("Middleware Kube Agent Config Updater Version", agentVersion)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("could not run application", zap.Error(err))
	}
}
