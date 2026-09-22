package agent

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

// The agent management APIs take the API key as a header rather than a URL
// path segment. A key in the path returned a confusing 404 from the router
// when it was malformed. See AGE-533.
func TestNewAgentAPIRequest(t *testing.T) {
	const apiKey = "testAPIKey"

	tests := []struct {
		name            string
		method          string
		url             string
		body            []byte
		wantBody        string
		wantContentType string
	}{
		{
			name:   "get without body",
			method: http.MethodGet,
			url:    "http://example.com/api/v1/agent/restart-status?host_id=myhost",
		},
		{
			name:            "post with body",
			method:          http.MethodPost,
			url:             "http://example.com/api/v1/agent/tracking",
			body:            []byte(`{"status":"validate"}`),
			wantBody:        `{"status":"validate"}`,
			wantContentType: "application/json",
		},
		{
			name:            "put with body",
			method:          http.MethodPut,
			url:             "http://example.com/api/v1/agent/public/setting/config-groups/group/default",
			body:            []byte(`{"hostIds":["myhost"]}`),
			wantBody:        `{"hostIds":["myhost"]}`,
			wantContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := newAgentAPIRequest(tt.method, tt.url, apiKey, tt.body)
			require.NoError(t, err)

			assert.Equal(t, tt.method, req.Method)
			assert.Equal(t, tt.url, req.URL.String())

			// The key travels in the header, never in the path.
			assert.Equal(t, apiKey, req.Header.Get(apiKeyHeader))
			assert.NotContains(t, req.URL.Path, apiKey)

			assert.Equal(t, tt.wantContentType, req.Header.Get("Content-Type"))

			if tt.body == nil {
				assert.Nil(t, req.Body)
				return
			}

			got, err := io.ReadAll(req.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, string(got))
		})
	}
}

func TestNewAgentAPIRequestInvalidMethod(t *testing.T) {
	_, err := newAgentAPIRequest("in valid", "http://example.com", "testAPIKey", nil)
	assert.Error(t, err)
}

// TestApplyConfigClassToHostsPath pins the config groups route. bifrost#14087
// inserted a literal "group" segment into it to work around a gin router
// conflict, which the earlier bifrost#14086 description does not show.
func TestApplyConfigClassToHostsPath(t *testing.T) {
	const apiKey = "testAPIKey"

	var gotHeader, gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get(apiKeyHeader)
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":true}`))
	}))
	defer srv.Close()

	cfg := HostConfig{
		BaseConfig: BaseConfig{
			APIKey:               apiKey,
			APIURLForConfigCheck: srv.URL,
		},
	}
	cfg.ConfigCheckInterval = "1s"

	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	require.NoError(t, agent.applyConfigClassToHosts())

	assert.Equal(t, http.MethodPut, gotMethod)
	assert.Equal(t, apiKey, gotHeader)
	assert.Equal(t, "/"+apiPathForConfigGroups+"/group/default", gotPath)
	assert.NotContains(t, gotPath, apiKey)
}

// TestUpdateConfigFileSendsAPIKeyHeader checks that the ingestion rules call
// reaches the backend with the key in the header and no key in the path.
func TestUpdateConfigFileSendsAPIKeyHeader(t *testing.T) {
	const apiKey = "testAPIKey"

	cfg := HostConfig{
		BaseConfig: BaseConfig{
			APIKey:               apiKey,
			APIURLForConfigCheck: "http://example.com",
		},
	}
	cfg.ConfigCheckInterval = "1s"

	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)

	var got *http.Request
	agent.httpDoFunc = func(req *http.Request) (*http.Response, error) {
		got = req
		// The body is irrelevant here; updateConfigFile is expected to fail
		// parsing it. This test only covers how the request is addressed.
		return nil, errors.New("test error")
	}

	_ = agent.updateConfigFile("nodocker")

	require.NotNil(t, got, "expected the ingestion rules api to be called")
	assert.Equal(t, apiKey, got.Header.Get(apiKeyHeader))
	assert.Equal(t, "/"+apiPathForYAML, got.URL.Path)
	assert.NotContains(t, got.URL.Path, apiKey)
}
