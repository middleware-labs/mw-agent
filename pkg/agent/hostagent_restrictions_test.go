package agent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

// restrictionsTestRules has the shape bifrost sends agents >= 1.21: renamed
// receivers (file_log, host_metrics) and exporter (otlp_grpc).
const restrictionsTestRules = `{"status":true,"config":{"nodocker":{
	"receivers":{
		"otlp":{"protocols":{"grpc":{"endpoint":"127.0.0.1:9319"}}},
		"file_log":{"include":["/var/log/*.log"],"start_at":"end"},
		"host_metrics":{"collection_interval":"60s","scrapers":{"cpu":{}}}},
	"processors":{"batch":{}},
	"exporters":{"otlp_grpc/2":{"endpoint":"127.0.0.1:4317"}},
	"service":{"pipelines":{
		"logs":{"receivers":["otlp","file_log"],"processors":["batch"],"exporters":["otlp_grpc/2"]},
		"metrics":{"receivers":["otlp","host_metrics"],"processors":["batch"],"exporters":["otlp_grpc/2"]}}}}}}`

func newRestrictionsTestAgent(t *testing.T, logs, metrics bool, apiURL string) *HostAgent {
	t.Helper()
	cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k", APIURLForConfigCheck: apiURL}}
	cfg.ConfigCheckInterval = "1s"
	cfg.OtelConfigFile = filepath.Join(t.TempDir(), "otel-config.yaml")
	cfg.AgentFeatures.LogCollection = logs
	cfg.AgentFeatures.MetricCollection = metrics
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)
	return agent
}

// With log or metric collection disabled, the rendered config drops the
// receivers under their new names too, not only the pipelines.
func TestRestrictionsRemoveRenamedReceiversFromRender(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(restrictionsTestRules))
	}))
	t.Cleanup(srv.Close)

	tests := []struct {
		name          string
		logs, metrics bool
		wantReceivers []string
		wantPipelines []string
	}{
		{"all enabled", true, true, []string{"file_log", "host_metrics", "otlp"}, []string{"logs", "metrics"}},
		{"logs disabled", false, true, []string{"host_metrics", "otlp"}, []string{"metrics"}},
		{"metrics disabled", true, false, []string{"file_log", "otlp"}, []string{"logs"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := newRestrictionsTestAgent(t, tt.logs, tt.metrics, srv.URL)
			path, err := agent.getOtelConfig()
			require.NoError(t, err)
			b, err := os.ReadFile(path)
			require.NoError(t, err)

			var rendered struct {
				Receivers map[string]interface{} `yaml:"receivers"`
				Service   struct {
					Pipelines map[string]interface{} `yaml:"pipelines"`
				} `yaml:"service"`
			}
			require.NoError(t, yaml.Unmarshal(b, &rendered))
			assert.ElementsMatch(t, tt.wantReceivers, keysOf(rendered.Receivers))
			assert.ElementsMatch(t, tt.wantPipelines, keysOf(rendered.Service.Pipelines))
		})
	}
}

// Both the old and the new name of every restricted receiver are deleted.
func TestUpdateConfigWithRestrictionsDeletesOldAndNewNames(t *testing.T) {
	logReceivers := []string{"filelog", "file_log", "windowseventlog"}
	metricReceivers := []string{"hostmetrics", "host_metrics", "windowsperfcounters",
		"docker_stats", "prometheus", "kubeletstats", "kubelet_stats", "k8s_cluster"}

	newConfig := func() map[string]interface{} {
		receivers := map[string]interface{}{"otlp": map[string]interface{}{}}
		for _, r := range append(append([]string{}, logReceivers...), metricReceivers...) {
			receivers[r] = map[string]interface{}{}
		}
		return map[string]interface{}{
			Receivers: receivers,
			Service:   map[string]interface{}{Pipelines: map[string]interface{}{}},
		}
	}

	tests := []struct {
		name          string
		logs, metrics bool
		gone          []string
	}{
		{"logs disabled", false, true, logReceivers},
		{"metrics disabled", true, false, metricReceivers},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := newRestrictionsTestAgent(t, tt.logs, tt.metrics, "http://127.0.0.1:1")
			config, err := agent.updateConfigWithRestrictions(newConfig())
			require.NoError(t, err)
			receivers := config[Receivers].(map[string]interface{})
			for _, r := range tt.gone {
				assert.NotContains(t, receivers, r)
			}
			// otlp plus the other feature's receivers stay.
			assert.Len(t, receivers, 1+len(logReceivers)+len(metricReceivers)-len(tt.gone))
		})
	}
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
