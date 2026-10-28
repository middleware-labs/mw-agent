package agent

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap/zapcore"
)

// restartTestRules is a minimal valid ingestion-rules answer.
const restartTestRules = `{"status":true,"config":{"nodocker":{
	"receivers":{"otlp":{"protocols":{"grpc":{"endpoint":"127.0.0.1:9319"}}}},
	"processors":{"batch":{}},
	"exporters":{"debug":{}},
	"service":{"pipelines":{"traces":{"receivers":["otlp"],"processors":["batch"],"exporters":["debug"]}}}}}}`

// A failed restart-status call is reported as ErrRestartStatusCheckFailure, so
// program.run keeps the collector in its current state. A good answer behaves
// as before.
func TestHostAgentCallRestartStatusAPI(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		closed  bool // the server is closed before the call: connection refused
		wantErr error
	}{
		{
			name:    "connection refused",
			closed:  true,
			wantErr: ErrRestartStatusCheckFailure,
		},
		{
			name: "502",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadGateway)
			},
			wantErr: ErrRestartStatusCheckFailure,
		},
		{
			name: "garbage body",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("<html>not json"))
			},
			wantErr: ErrRestartStatusCheckFailure,
		},
		{
			name: "restart true fetches the config and restarts",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, apiPathForYAML) {
					w.Write([]byte(restartTestRules))
					return
				}
				w.Write([]byte(`{"status":true,"restart":true}`))
			},
			wantErr: ErrRestartAgent,
		},
		{
			name: "restart false",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"status":true,"restart":false}`))
			},
		},
		{
			// What bifrost answers for an invalid token: still 200.
			name: "status false",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"status":false,"message":"Invalid account agent token","restart":false}`))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			if tt.closed {
				srv.Close()
			} else {
				t.Cleanup(srv.Close)
			}
			cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k", APIURLForConfigCheck: srv.URL}}
			cfg.ConfigCheckInterval = "1s"
			cfg.OtelConfigFile = filepath.Join(t.TempDir(), "otel-config.yaml")
			cfg.AgentFeatures.LogCollection = true
			cfg.AgentFeatures.MetricCollection = true
			agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
			require.NoError(t, err)

			err = agent.callRestartStatusAPI()
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantErr == ErrRestartStatusCheckFailure {
				assert.NotErrorIs(t, err, ErrRestartAgent)
			}
		})
	}
}

// col_running on the restart-status and ingestion-rules calls is 1 while the
// collector runs and 0 otherwise (it was inverted).
func TestColRunningParam(t *testing.T) {
	for _, running := range []bool{false, true} {
		t.Run(map[bool]string{false: "stopped", true: "running"}[running], func(t *testing.T) {
			got := map[string]string{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got[r.URL.Path] = r.URL.Query().Get("col_running")
				if strings.Contains(r.URL.Path, apiPathForRestart) {
					w.Write([]byte(`{"status":true,"restart":false}`))
					return
				}
				w.Write([]byte(restartTestRules))
			}))
			t.Cleanup(srv.Close)
			cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k", APIURLForConfigCheck: srv.URL}}
			cfg.ConfigCheckInterval = "1s"
			cfg.OtelConfigFile = filepath.Join(t.TempDir(), "otel-config.yaml")
			agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
			require.NoError(t, err)
			if running {
				agent.collector, err = otelcol.NewCollector(agent.collectorSettings)
				require.NoError(t, err)
			}

			require.NoError(t, agent.callRestartStatusAPI())
			_, err = agent.fetchIngestionRules("nodocker")
			require.NoError(t, err)

			want := map[bool]string{false: "0", true: "1"}[running]
			assert.Equal(t, map[string]string{"/" + apiPathForRestart: want, "/" + apiPathForYAML: want}, got)
		})
	}
}
