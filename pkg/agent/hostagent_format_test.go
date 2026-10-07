package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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
// In supervised mode the agent still polls restart-status itself (as a
// standalone agent does), but on a change it hands the raw ingestion rules to
// the OpAMP server instead of writing otel-config.yaml or restarting in-process.
// A failed push keeps the collector running and is retried on the next tick,
// because the backend clears its restart flag on the first read.
func TestListenForConfigChangesSupervisedMode(t *testing.T) {
	const rules = `{"status":true,"config":{"nodocker":{"receivers":{}}}}`

	var mu sync.Mutex
	var backendPaths []string
	restartCalls := 0
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		backendPaths = append(backendPaths, r.URL.Path)
		mu.Unlock()
		switch {
		case strings.Contains(r.URL.Path, "restart-status"):
			mu.Lock()
			restartCalls++
			first := restartCalls == 1
			mu.Unlock()
			_, _ = fmt.Fprintf(w, `{"status":true,"restart":%t}`, first)
		case strings.Contains(r.URL.Path, "ingestion-rules"):
			_, _ = io.WriteString(w, rules)
		default:
			_, _ = io.WriteString(w, `{"status":true}`)
		}
	}))
	defer backend.Close()

	type push struct{ path, contentType, auth, body string }
	pushes := make(chan push, 8)
	var pushCount atomic.Int32
	restartCallsAtRetry := -1
	opamp := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		pushes <- push{r.URL.Path, r.Header.Get("Content-Type"), r.Header.Get("Authorization"), string(b)}
		if pushCount.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable) // first push fails
			return
		}
		mu.Lock()
		restartCallsAtRetry = restartCalls
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer opamp.Close()

	otelFile := filepath.Join(t.TempDir(), "otel-config.yaml")
	cfg := HostConfig{BaseConfig: BaseConfig{
		APIKey:               "testAPIKey",
		APIURLForConfigCheck: backend.URL,
		RemoteAgentEnabled:   true,
		AgentID:              "0198a6d2-0000-7000-8000-000000000001",
		OpAMPServerURL:       strings.Replace(opamp.URL, "http://", "ws://", 1) + "/v1/opamp",
		OtelConfigFile:       otelFile,
	}}
	cfg.ConfigCheckInterval = "50ms"
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	errCh := make(chan error, 16)
	stopCh := make(chan struct{})
	done := make(chan struct{})
	go func() { _ = agent.ListenForConfigChanges(errCh, stopCh); close(done) }()

	assert.NoError(t, <-errCh, "supervised mode reports ready without fetching config")
	for i := 0; i < 2; i++ {
		select {
		case p := <-pushes:
			assert.Equal(t, "/api/v1/agents/0198a6d2-0000-7000-8000-000000000001/config/push", p.path)
			assert.Equal(t, "application/json", p.contentType)
			assert.Equal(t, "Bearer testAPIKey", p.auth)
			assert.JSONEq(t, rules, p.body, "raw ingestion rules are pushed unrendered")
		case <-time.After(2 * time.Second):
			t.Fatalf("push %d to the OpAMP server never arrived", i+1)
		}
		err := <-errCh
		if i == 0 {
			assert.ErrorIs(t, err, ErrConfigFetchFailure, "a failed push leaves the collector running")
		} else {
			assert.NoError(t, err, "the retried push succeeds without an in-process restart")
		}
	}
	close(stopCh)
	<-done

	_, statErr := os.Stat(otelFile)
	assert.True(t, os.IsNotExist(statErr), "supervised mode never writes otel-config.yaml")

	mu.Lock()
	defer mu.Unlock()
	groupCalls := 0
	for _, p := range backendPaths {
		if strings.Contains(p, "/config-groups/group/default") {
			groupCalls++
		}
	}
	assert.Equal(t, 1, groupCalls, "config-group registration runs once, got %v", backendPaths)
	assert.Equal(t, 1, restartCallsAtRetry, "the retry tick re-pushes without polling restart-status again")
}
