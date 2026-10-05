package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

// BuildOtelConfig is what `mw-agent format` runs for the OpAMP supervisor. It must
// merge the on-host credential file named by the backend response into the
// matching receiver and return YAML that validates against this binary.
func TestBuildOtelConfigMergesCredentialFile(t *testing.T) {
	t.Setenv("MW_TARGET", "localhost:443")
	t.Setenv("MW_API_KEY", "k")

	backend := map[string]any{
		"status": true,
		"config": map[string]any{
			"nodocker": map[string]any{
				"receivers": map[string]any{
					"nop":   map[string]any{},
					"redis": map[string]any{"endpoint": "placeholder:1"},
				},
				"exporters": map[string]any{"nop": map[string]any{}},
				"service": map[string]any{
					"pipelines": map[string]any{
						"metrics": map[string]any{
							"receivers": []any{"redis", "nop"},
							"exporters": []any{"nop"},
						},
					},
				},
			},
		},
		// path is relative to the package dir, like the other credential tests
		"redis_config": map[string]any{"path": "db-config_test.yaml"},
	}
	body, err := json.Marshal(backend)
	require.NoError(t, err)

	cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k", Target: "localhost"}}
	cfg.ConfigCheckInterval = "1s"
	cfg.AgentFeatures.LogCollection = true
	cfg.AgentFeatures.MetricCollection = true
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	out, err := agent.BuildOtelConfig(body, "nodocker")
	require.NoError(t, err)

	var rendered map[string]any
	require.NoError(t, yaml.Unmarshal(out, &rendered))
	receivers := rendered["receivers"].(map[any]any)
	redis := receivers["redis"].(map[any]any)
	assert.Equal(t, "localhost:7379", redis["endpoint"], "credential file must win over the backend value")
	assert.Equal(t, "mypassword", redis["password"], "credential file keys must be merged in")
	assert.NotContains(t, string(out), "redis_config", "integration blocks must not leak into the collector config")

	// A response without any config is rejected.
	_, err = agent.BuildOtelConfig([]byte(`{"status":true,"config":{}}`), "nodocker")
	assert.Error(t, err)

	// Asking for the docker variant of a response that only carries nodocker
	// (the backend returns just the requested file) must render the one present,
	// not an empty config.
	out, err = agent.BuildOtelConfig(body, "docker")
	require.NoError(t, err)
	assert.Contains(t, string(out), "redis")
}

// In supervised mode the agent must not fetch config or poll restart-status (the
// OpAMP server does, and the backend clears the flag on first read). The only
// backend call left is the one-time config-group registration.
func TestListenForConfigChangesSupervisedMode(t *testing.T) {
	cfg := HostConfig{BaseConfig: BaseConfig{
		APIKey:               "testAPIKey",
		APIURLForConfigCheck: "http://example.com",
		RemoteAgentEnabled:   true,
	}}
	cfg.ConfigCheckInterval = "50ms"
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	var mu sync.Mutex
	var paths []string
	agent.httpDoFunc = func(req *http.Request) (*http.Response, error) {
		mu.Lock()
		paths = append(paths, req.URL.Path)
		mu.Unlock()
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"status":true}`))}, nil
	}

	errCh := make(chan error, 4)
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() { _ = agent.ListenForConfigChanges(errCh, stopCh); close(done) }()

	assert.NoError(t, <-errCh, "supervised mode reports ready without fetching config")
	time.Sleep(300 * time.Millisecond)
	close(stopCh)
	<-done

	mu.Lock()
	defer mu.Unlock()
	require.Len(t, paths, 1, "exactly one backend call (config-group registration), got %v", paths)
	assert.Contains(t, paths[0], "/config-groups/group/default")
	for _, p := range paths {
		assert.NotContains(t, p, "restart-status")
		assert.NotContains(t, p, "ingestion-rules")
	}
}
