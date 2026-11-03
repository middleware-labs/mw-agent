package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componentstatus"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const collectorFailedLine = "collector server run finished with error"

// A pipeline with this regex makes the collector fail to start, as on a real host.
const collectorConfigFormat = `
receivers:
  file_log:
    include: [/nonexistent/mw-agent-test.log]
    operators:
      - type: regex_parser
        regex: '%s'
exporters:
  debug: {}
service:
  telemetry:
    metrics:
      level: none
  pipelines:
    logs:
      receivers: [file_log]
      exporters: [debug]
`

func newCollectorExitAgent(t *testing.T, regex string, supervised bool) (*HostAgent, chan int) {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "otel-config.yaml")
	config := []byte(fmt.Sprintf(collectorConfigFormat, regex))
	require.NoError(t, os.WriteFile(configPath, config, 0o600))

	cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k"}}
	cfg.ConfigCheckInterval = "1s"
	cfg.OtelConfigFile = configPath
	cfg.RemoteAgentEnabled = supervised
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	exits := make(chan int, 1)
	agent.exit = func(code int) { exits <- code }
	return agent, exits
}

func TestCollectorFailureExitsInSupervisedMode(t *testing.T) {
	agent, exits := newCollectorExitAgent(t, "123333", true)

	require.NoError(t, agent.StartCollector())

	select {
	case code := <-exits:
		assert.Equal(t, exitCodeCollectorFailed, code)
	case <-time.After(60 * time.Second):
		t.Fatal("the agent did not exit after its collector failed")
	}
}

// A standalone agent gets a new config from the backend and restarts its
// collector itself, so it must stay up.
func TestCollectorFailureKeepsStandaloneAgentRunning(t *testing.T) {
	agent, exits := newCollectorExitAgent(t, "123333", false)

	require.NoError(t, agent.StartCollector())
	agent.collectorWG.Wait()

	assert.Empty(t, exits, "a standalone agent must not exit when its collector fails")
}

// The supervisor stops the collector on purpose too; that is not a failure.
func TestCollectorStoppedOnPurposeDoesNotExit(t *testing.T) {
	agent, exits := newCollectorExitAgent(t, `(?P<level>\w+)`, true)

	require.NoError(t, agent.StartCollector())
	collector := agent.collector
	require.NotNil(t, collector)
	require.Eventually(t, func() bool {
		return collector.GetState() == otelcol.StateRunning
	}, 60*time.Second, 50*time.Millisecond, "collector did not start")

	agent.StopCollector(nil)

	assert.Empty(t, exits, "stopping the collector on purpose must not exit the agent")
}

// shutdownFailingExtension returns an error from Shutdown.
type shutdownFailingExtension struct {
	component.StartFunc
}

func (e *shutdownFailingExtension) Shutdown(context.Context) error {
	return errors.New("drain failed")
}

// fatalReceiver reports a fatal error after it has started, as a component that
// fails at runtime does.
type fatalReceiver struct {
	component.ShutdownFunc
}

func (r *fatalReceiver) Start(_ context.Context, host component.Host) error {
	time.AfterFunc(500*time.Millisecond, func() {
		componentstatus.ReportStatus(host, componentstatus.NewFatalErrorEvent(errors.New("listener died")))
	})
	return nil
}

const testComponentsConfig = `
extensions:
  shutdown_failing: {}
receivers:
  fatal_receiver: {}
  file_log:
    include: [/nonexistent/mw-agent-test.log]
exporters:
  debug: {}
service:
  extensions: [%s]
  telemetry:
    metrics:
      level: none
  pipelines:
    logs:
      receivers: [%s]
      exporters: [debug]
`

// newAgentWithTestComponents returns a supervised or standalone agent whose
// collector runs the given extensions and receiver, the logs it writes, and the
// exit codes it asks for.
func newAgentWithTestComponents(t *testing.T, extensions, receiverName string, supervised bool) (*HostAgent, *observer.ObservedLogs, chan int) {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "otel-config.yaml")
	config := fmt.Sprintf(testComponentsConfig, extensions, receiverName)
	require.NoError(t, os.WriteFile(configPath, []byte(config), 0o600))

	cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k"}}
	cfg.ConfigCheckInterval = "1s"
	cfg.OtelConfigFile = configPath
	cfg.RemoteAgentEnabled = supervised
	core, logs := observer.New(zapcore.InfoLevel)
	agent, err := NewHostAgent(cfg, core)
	require.NoError(t, err)
	agent.logger = zap.New(core)

	getFactories := agent.collectorSettings.Factories
	agent.collectorSettings.Factories = func() (otelcol.Factories, error) {
		factories, err := getFactories()
		if err != nil {
			return factories, err
		}
		emptyConfig := func() component.Config { return &struct{}{} }
		extensionType := component.MustNewType("shutdown_failing")
		factories.Extensions[extensionType] = extension.NewFactory(extensionType, emptyConfig,
			func(context.Context, extension.Settings, component.Config) (extension.Extension, error) {
				return &shutdownFailingExtension{}, nil
			},
			component.StabilityLevelDevelopment)
		receiverType := component.MustNewType("fatal_receiver")
		factories.Receivers[receiverType] = receiver.NewFactory(receiverType, emptyConfig,
			receiver.WithLogs(func(context.Context, receiver.Settings, component.Config, consumer.Logs) (receiver.Logs, error) {
				return &fatalReceiver{}, nil
			}, component.StabilityLevelDevelopment))
		return factories, nil
	}

	exits := make(chan int, 1)
	agent.exit = func(code int) { exits <- code }
	return agent, logs, exits
}

// Run also returns a component's shutdown error after a stop on purpose. That is
// no failure: no exit, and not the line the supervisor reports as the reason.
func TestCollectorShutdownErrorOnPurposeIsNoFailure(t *testing.T) {
	agent, logs, exits := newAgentWithTestComponents(t, "shutdown_failing", "file_log", true)
	require.NoError(t, agent.StartCollector())
	collector := agent.runningCollector()
	require.NotNil(t, collector)
	require.Eventually(t, func() bool {
		return collector.GetState() == otelcol.StateRunning
	}, 60*time.Second, 50*time.Millisecond, "collector did not start")

	agent.StopCollector(nil)

	assert.Empty(t, exits)
	assert.Zero(t, logs.FilterMessage(collectorFailedLine).Len())
	assert.Equal(t, 1, logs.FilterMessage("collector stopped on request, with a shutdown error").Len())
	assert.Nil(t, agent.runningCollector())
}

// After a component's fatal error Run returns nil, but the collector is gone.
func TestCollectorFatalErrorExitsInSupervisedMode(t *testing.T) {
	agent, logs, exits := newAgentWithTestComponents(t, "", "fatal_receiver", true)

	require.NoError(t, agent.StartCollector())

	select {
	case code := <-exits:
		assert.Equal(t, exitCodeCollectorFailed, code)
	case <-time.After(60 * time.Second):
		t.Fatal("the agent did not exit after its collector stopped")
	}
	assert.Equal(t, 1, logs.FilterMessage(collectorFailedLine).Len())
}

func TestCollectorFatalErrorKeepsStandaloneAgentRunning(t *testing.T) {
	agent, logs, exits := newAgentWithTestComponents(t, "", "fatal_receiver", false)

	require.NoError(t, agent.StartCollector())
	agent.collectorWG.Wait()

	assert.Empty(t, exits)
	assert.Zero(t, logs.FilterMessage(collectorFailedLine).Len())
}
