package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap/zapcore"
)

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
