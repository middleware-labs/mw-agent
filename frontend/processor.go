package frontend

import (
	"context"
	"errors"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"log"
)

// MetaProcessor is an implementation of configauth.GRPCClientAuthenticator. It embeds a static authorization "bearer" token in every rpc call.
type MetaProcessor struct {
	config *ConfigProcessor
	logger *zap.Logger
}

func (b *MetaProcessor) ConsumeLogs(ctx context.Context, ld plog.Logs) error {
	//TODO implement me
	panic("implement me ConsumeLogs")
}

func (b *MetaProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	//TODO implement me
	panic("implement me ConsumeMetrics")
}

func (b *MetaProcessor) Capabilities() consumer.Capabilities {
	//TODO implement me
	panic("implement me Capabilities")
}

func (b *MetaProcessor) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	//TODO implement me
	panic("implement me ConsumeTraces")
}

// Start of MetaProcessor does nothing and returns nil
func (b *MetaProcessor) Start(ctx context.Context, host component.Host) error {
	return nil
}

// Shutdown of MetaProcessor does nothing and returns nil
func (b *MetaProcessor) Shutdown(ctx context.Context) error {
	return nil
}

var (
	_ component.TracesProcessor  = (*MetaProcessor)(nil)
	_ component.MetricsProcessor = (*MetaProcessor)(nil)
	_ component.LogsProcessor    = (*MetaProcessor)(nil)

	errAuthenticationRequired = errors.New("authentication required")
)

// Config specifies how the Per-RPC bearer token based authentication data should be obtained.
type ConfigProcessor struct {
	config.ProcessorSettings `mapstructure:",squash"`

	// BearerToken specifies the bearer token to use for every RPC.
	Server string `mapstructure:"server,omitempty"`
	Token  string `mapstructure:"token,omitempty"`
}

var _ config.Processor = (*ConfigProcessor)(nil)

// Validate checks if the extension configuration is valid
func (cfg *ConfigProcessor) Validate() error {
	log.Println("NewProcessorFactory ValidateValidateValidate " + cfg.Server)
	log.Println("NewProcessorFactory ValidateValidateValidate " + cfg.Token)
	if cfg.Server == "" || cfg.Token == "" {
		log.Println("NewAuthFactory error " + errNoTokenProvided.Error())
		return errNoTokenProvided
	}
	return nil
}
