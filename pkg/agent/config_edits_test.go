package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func newFormatTestAgent(t *testing.T, mutate func(*HostConfig)) *HostAgent {
	t.Helper()
	t.Setenv("MW_TARGET", "localhost:443")
	t.Setenv("MW_API_KEY", "k")
	cfg := HostConfig{BaseConfig: BaseConfig{APIKey: "k", Target: "localhost"}}
	cfg.ConfigCheckInterval = "1s"
	cfg.AgentFeatures.LogCollection = true
	cfg.AgentFeatures.MetricCollection = true
	if mutate != nil {
		mutate(&cfg)
	}
	agent, err := NewHostAgent(cfg, zapcore.NewNopCore())
	require.NoError(t, err)
	return agent
}

// withTrackCounter points the agent's backend at a server that counts
// UpdateAgentTrackStatus calls.
func withTrackCounter(t *testing.T, calls *atomic.Int32) func(*HostConfig) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, apiAgentTrack) {
			calls.Add(1)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return func(cfg *HostConfig) { cfg.APIURLForConfigCheck = srv.URL }
}

func readFormatFixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "format", name))
	require.NoError(t, err)
	return b
}

func parseRendered(t *testing.T, b []byte) map[interface{}]interface{} {
	t.Helper()
	var m map[interface{}]interface{}
	require.NoError(t, yaml.Unmarshal(b, &m))
	return m
}

func yamlAt(m interface{}, keys ...string) interface{} {
	for _, k := range keys {
		mm, ok := m.(map[interface{}]interface{})
		if !ok {
			return nil
		}
		m = mm[k]
	}
	return m
}

func render(t *testing.T, agent *HostAgent, body []byte, edits string) *ConfigRender {
	t.Helper()
	r, err := agent.RenderOtelConfig(body, "nodocker", []byte(edits))
	require.NoError(t, err)
	require.NotNil(t, r.Status)
	return r
}

func opResults(ops []EditOpResult) []string {
	out := make([]string, 0, len(ops))
	for _, op := range ops {
		out = append(out, op.Result)
	}
	return out
}

// Without edits the base is exactly what BuildOtelConfig renders, and both are
// byte-identical to the output of the code before RenderOtelConfig existed
// (golden files generated with mw-agent cd59138). Run with -count=20.
func TestRenderOtelConfigGoldenWithoutEdits(t *testing.T) {
	variants := map[string]func(*HostConfig){
		"plain":    nil,
		"hosttags": func(c *HostConfig) { c.HostTags = "env:prod,team:core" },
		"nologs":   func(c *HostConfig) { c.AgentFeatures.LogCollection = false },
	}
	for _, fixture := range []string{"host", "redis", "postgres"} {
		for variant, mutate := range variants {
			t.Run(fixture+"/"+variant, func(t *testing.T) {
				agent := newFormatTestAgent(t, mutate)
				body := readFormatFixture(t, fixture+".json")

				built, buildErr := agent.BuildOtelConfig(body, "nodocker")
				r, renderErr := agent.RenderOtelConfig(body, "nodocker", nil)

				if wantErr, err := os.ReadFile(filepath.Join("testdata", "format", fixture+"."+variant+".err")); err == nil {
					require.EqualError(t, buildErr, string(wantErr))
					require.EqualError(t, renderErr, string(wantErr))
					return
				}
				golden := readFormatFixture(t, fixture+"."+variant+".golden.yaml")
				require.NoError(t, buildErr)
				require.NoError(t, renderErr)
				assert.Equal(t, string(golden), string(built))
				assert.Equal(t, string(golden), string(r.Base))
				assert.Equal(t, string(golden), string(r.Final))
				assert.Nil(t, r.Status)

				// An empty edits document changes nothing either.
				empty := render(t, agent, body, `[]`)
				assert.Equal(t, string(golden), string(empty.Final))
				assert.True(t, empty.Status.Applied)
				assert.Empty(t, empty.Status.Ops)
			})
		}
	}
}

// One case per row of "Exact edit rules" and each result it can give.
func TestEditRules(t *testing.T) {
	cases := []struct {
		name    string
		edits   string
		want    []string
		reasons []string // substrings, per entry; "" = no check
		check   func(t *testing.T, final map[interface{}]interface{})
	}{
		{
			name:  "add new component",
			edits: `[{"op":"add","path":["receivers","otlp/extra"],"value":{"protocols":{"http":{"endpoint":"127.0.0.1:4319"}}}}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, "127.0.0.1:4319", yamlAt(f, "receivers", "otlp/extra", "protocols", "http", "endpoint"))
			},
		},
		{
			name:    "add existing component is a conflict and the edit wins",
			edits:   `[{"op":"add","path":["receivers","otlp"],"value":{"protocols":{"grpc":{"endpoint":"127.0.0.1:9400"}}}}]`,
			want:    []string{EditConflict},
			reasons: []string{"already in base"},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, "127.0.0.1:9400", yamlAt(f, "receivers", "otlp", "protocols", "grpc", "endpoint"))
			},
		},
		{
			name:  "add pipeline",
			edits: `[{"op":"add","path":["service","pipelines","logs/app"],"value":{"receivers":["otlp"],"exporters":["otlp_grpc/2"]}}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, []interface{}{"otlp"}, yamlAt(f, "service", "pipelines", "logs/app", "receivers"))
			},
		},
		{
			name: "remove component and take it out of its pipeline",
			edits: `[{"op":"remove","path":["receivers","file_log"]},
			         {"op":"list_remove","path":["service","pipelines","logs","receivers"],"value":"file_log"}]`,
			want: []string{EditApplied, EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Nil(t, yamlAt(f, "receivers", "file_log"))
				assert.Equal(t, []interface{}{"otlp"}, yamlAt(f, "service", "pipelines", "logs", "receivers"))
			},
		},
		{
			name:    "remove absent component is skipped",
			edits:   `[{"op":"remove","path":["receivers","nope"]}]`,
			want:    []string{EditSkipped},
			reasons: []string{"component receivers/nope not in base"},
		},
		{
			name:  "set a setting",
			edits: `[{"op":"set","path":["processors","batch","timeout"],"value":"2s"}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, "2s", yamlAt(f, "processors", "batch", "timeout"))
			},
		},
		{
			name:  "set creates missing maps inside the component",
			edits: `[{"op":"set","path":["receivers","otlp","protocols","http","endpoint"],"value":"127.0.0.1:4318"}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, "127.0.0.1:4318", yamlAt(f, "receivers", "otlp", "protocols", "http", "endpoint"))
				assert.Equal(t, "127.0.0.1:9319", yamlAt(f, "receivers", "otlp", "protocols", "grpc", "endpoint"))
			},
		},
		{
			name:  "set replaces a list inside a component as one value",
			edits: `[{"op":"set","path":["receivers","file_log","include"],"value":["/srv/*.log"]}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, []interface{}{"/srv/*.log"}, yamlAt(f, "receivers", "file_log", "include"))
			},
		},
		{
			name: "set on an absent component is skipped, never half-created",
			edits: `[{"op":"set","path":["processors","memory_limiter","check_interval"],"value":"1s"},
			         {"op":"set","path":["service","telemetry","logs","level"],"value":"debug"}]`,
			want:    []string{EditSkipped, EditSkipped},
			reasons: []string{"component processors/memory_limiter not in base", "component service/telemetry not in base"},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Nil(t, yamlAt(f, "processors", "memory_limiter"))
				assert.Nil(t, yamlAt(f, "service", "telemetry"))
			},
		},
		{
			name:    "set across a non-map is skipped",
			edits:   `[{"op":"set","path":["receivers","file_log","include","first"],"value":"x"}]`,
			want:    []string{EditSkipped},
			reasons: []string{"crosses a non-map"},
		},
		{
			name:  "unset a setting",
			edits: `[{"op":"unset","path":["processors","batch","send_batch_size"]}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.NotContains(t, yamlAt(f, "processors", "batch"), "send_batch_size")
				assert.Equal(t, "10s", yamlAt(f, "processors", "batch", "timeout"))
			},
		},
		{
			name: "unset of something absent is skipped",
			edits: `[{"op":"unset","path":["processors","batch","nope"]},
			         {"op":"unset","path":["processors","nope","timeout"]}]`,
			want:    []string{EditSkipped, EditSkipped},
			reasons: []string{"processors.batch.nope not in base", "component processors/nope not in base"},
		},
		{
			name: "list_add to service.extensions",
			edits: `[{"op":"add","path":["extensions","health_check"],"value":{}},
			         {"op":"list_add","path":["service","extensions"],"value":"health_check"}]`,
			want: []string{EditApplied, EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, []interface{}{"health_check"}, yamlAt(f, "service", "extensions"))
			},
		},
		{
			name: "list_add to pipeline exporters",
			edits: `[{"op":"add","path":["exporters","debug"],"value":{}},
			         {"op":"list_add","path":["service","pipelines","logs","exporters"],"value":"debug"}]`,
			want: []string{EditApplied, EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, []interface{}{"otlp_grpc/2", "debug"}, yamlAt(f, "service", "pipelines", "logs", "exporters"))
			},
		},
		{
			name: "list_add already there or to an absent pipeline is skipped",
			edits: `[{"op":"list_add","path":["service","pipelines","logs","receivers"],"value":"otlp"},
			         {"op":"list_add","path":["service","pipelines","traces","receivers"],"value":"otlp"}]`,
			want:    []string{EditSkipped, EditSkipped},
			reasons: []string{"already in", "component service/pipelines/traces not in base"},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Nil(t, yamlAt(f, "service", "pipelines", "traces"))
			},
		},
		{
			name:    "list_remove of a name not there is skipped",
			edits:   `[{"op":"list_remove","path":["service","pipelines","metrics","receivers"],"value":"file_log"}]`,
			want:    []string{EditSkipped},
			reasons: []string{"file_log not in service.pipelines.metrics.receivers"},
		},
		{
			name:  "list_set replaces the processors",
			edits: `[{"op":"list_set","path":["service","pipelines","metrics","processors"],"value":["batch","resourcedetection"]}]`,
			want:  []string{EditApplied},
			check: func(t *testing.T, f map[interface{}]interface{}) {
				assert.Equal(t, []interface{}{"batch", "resourcedetection"}, yamlAt(f, "service", "pipelines", "metrics", "processors"))
			},
		},
		{
			name:    "list_set on an absent pipeline is skipped",
			edits:   `[{"op":"list_set","path":["service","pipelines","traces","processors"],"value":["batch"]}]`,
			want:    []string{EditSkipped},
			reasons: []string{"component service/pipelines/traces not in base"},
		},
	}

	body := readFormatFixture(t, "host.json")
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			agent := newFormatTestAgent(t, nil)
			r := render(t, agent, body, tc.edits)
			require.Empty(t, r.Status.Error)
			assert.True(t, r.Status.Applied)
			assert.Equal(t, tc.want, opResults(r.Status.Ops))
			for i, want := range tc.reasons {
				if want != "" {
					assert.Contains(t, r.Status.Ops[i].Reason, want)
				}
			}
			for i, op := range r.Status.Ops {
				assert.Equal(t, i, op.I)
				if op.Result == EditApplied {
					assert.Empty(t, op.Reason)
				}
			}
			if tc.check != nil {
				tc.check(t, parseRendered(t, r.Final))
			}
		})
	}
}

// Owned paths and wrong list ops are refused one by one; the other entries
// still apply.
func TestEditRefusedPaths(t *testing.T) {
	refused := map[string]string{
		"extensions.opamp setting":    `{"op":"set","path":["extensions","opamp","server","ws","endpoint"],"value":"ws://evil"}`,
		"extensions.opamp add":        `{"op":"add","path":["extensions","opamp"],"value":{}}`,
		"extensions.opamp remove":     `{"op":"remove","path":["extensions","opamp"]}`,
		"opamp into service.ext":      `{"op":"list_add","path":["service","extensions"],"value":"opamp"}`,
		"opamp out of service.ext":    `{"op":"list_remove","path":["service","extensions"],"value":"opamp"}`,
		"instance id":                 `{"op":"set","path":["service","telemetry","resource","service.instance.id"],"value":"x"}`,
		"host_tags processor setting": `{"op":"set","path":["processors","resource/host_tags","attributes"],"value":[]}`,
		"host_tags processor remove":  `{"op":"remove","path":["processors","resource/host_tags"]}`,
		"list_add on processors":      `{"op":"list_add","path":["service","pipelines","logs","processors"],"value":"batch"}`,
		"list_set on receivers":       `{"op":"list_set","path":["service","pipelines","logs","receivers"],"value":["otlp"]}`,
		"list_set on service.ext":     `{"op":"list_set","path":["service","extensions"],"value":[]}`,
		"set on a pipeline list":      `{"op":"set","path":["service","pipelines","logs","receivers"],"value":["otlp"]}`,
		"unset on a pipeline list":    `{"op":"unset","path":["service","pipelines","logs","processors"]}`,
		"remove service.extensions":   `{"op":"remove","path":["service","extensions"]}`,
	}
	body := readFormatFixture(t, "host.json")
	for name, entry := range refused {
		t.Run(name, func(t *testing.T) {
			agent := newFormatTestAgent(t, func(c *HostConfig) { c.HostTags = "env:prod" })
			r := render(t, agent, body, `[`+entry+`,{"op":"set","path":["processors","batch","timeout"],"value":"2s"}]`)
			require.Empty(t, r.Status.Error)
			assert.True(t, r.Status.Applied)
			assert.Equal(t, []string{EditRefused, EditApplied}, opResults(r.Status.Ops))
			assert.NotEmpty(t, r.Status.Ops[0].Reason)

			// The final is the base plus only the timeout change.
			want := parseRendered(t, r.Base)
			yamlAt(want, "processors", "batch").(map[interface{}]interface{})["timeout"] = "2s"
			assert.Equal(t, want, parseRendered(t, r.Final))
		})
	}
}

// resource/host_tags in a list_set value is dropped silently and the host-tags
// step puts it back in front.
func TestEditListSetDropsHostTagsProcessor(t *testing.T) {
	agent := newFormatTestAgent(t, func(c *HostConfig) { c.HostTags = "env:prod" })
	body := readFormatFixture(t, "host.json")
	r := render(t, agent, body, `[{"op":"list_set","path":["service","pipelines","metrics","processors"],"value":["batch","resource/host_tags","resourcedetection"]}]`)
	assert.Equal(t, []string{EditApplied}, opResults(r.Status.Ops))
	assert.Equal(t, []interface{}{"resource/host_tags", "batch", "resourcedetection"},
		yamlAt(parseRendered(t, r.Final), "service", "pipelines", "metrics", "processors"))
}

// A malformed document applies nothing: final = base, applied false, no ops.
func TestEditsMalformedDocument(t *testing.T) {
	const ok = `{"op":"set","path":["processors","batch","timeout"],"value":"2s"}`
	cases := map[string]string{
		"not an array":             `{"op":"set"}`,
		"string document":          `"[]"`,
		"empty file":               ``,
		"null":                     `null`,
		"unknown op":               `[` + ok + `,{"op":"bogus","path":["receivers","otlp"]}]`,
		"empty path":               `[` + ok + `,{"op":"remove","path":[]}]`,
		"empty key in path":        `[` + ok + `,{"op":"remove","path":["receivers",""]}]`,
		"unknown section":          `[` + ok + `,{"op":"set","path":["bogus","x","y"],"value":1}]`,
		"entry not an object":      `[` + ok + `,1]`,
		"path not strings":         `[` + ok + `,{"op":"remove","path":["receivers",1]}]`,
		"add not a component path": `[` + ok + `,{"op":"add","path":["receivers","otlp","protocols"],"value":{}}]`,
		"remove of a section":      `[` + ok + `,{"op":"remove","path":["receivers"]}]`,
		"set on a component path":  `[` + ok + `,{"op":"set","path":["processors","batch"],"value":{}}]`,
		"pipelines without a name": `[` + ok + `,{"op":"set","path":["service","pipelines"],"value":{}}]`,
		"list op on a non-list":    `[` + ok + `,{"op":"list_add","path":["receivers","file_log","include"],"value":"x"}]`,
		"add value not an object":  `[` + ok + `,{"op":"add","path":["receivers","x"],"value":"y"}]`,
		"set without value":        `[` + ok + `,{"op":"set","path":["processors","batch","timeout"]}]`,
		"list_add value a number":  `[` + ok + `,{"op":"list_add","path":["service","extensions"],"value":1}]`,
		"list_set value a string":  `[` + ok + `,{"op":"list_set","path":["service","pipelines","logs","processors"],"value":"batch"}]`,
		"list_set value mixed":     `[` + ok + `,{"op":"list_set","path":["service","pipelines","logs","processors"],"value":["batch",2]}]`,
	}
	body := readFormatFixture(t, "host.json")
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			agent := newFormatTestAgent(t, nil)
			r := render(t, agent, body, doc)
			assert.Equal(t, string(r.Base), string(r.Final))
			assert.False(t, r.Status.Applied)
			assert.Contains(t, r.Status.Error, "malformed edits")
			assert.Empty(t, r.Status.Ops)
		})
	}
}

func TestEditsPanicFallsBackToBase(t *testing.T) {
	editStepTestHook = func() { panic("injected") }
	t.Cleanup(func() { editStepTestHook = nil })

	agent := newFormatTestAgent(t, nil)
	body := readFormatFixture(t, "host.json")
	r := render(t, agent, body, `[{"op":"set","path":["processors","batch","timeout"],"value":"2s"}]`)
	assert.Equal(t, string(r.Base), string(r.Final))
	assert.False(t, r.Status.Applied)
	assert.Equal(t, "edit step panicked: injected", r.Status.Error)
	assert.Empty(t, r.Status.Ops)
}

// The worked example from opamp-config-edits-explained §7: the user's four
// edits on base v1, then the same edits on base v2 (an integration added
// postgres). Checked as parsed YAML.
func TestEditsWorkedExample(t *testing.T) {
	baseV1 := map[string]interface{}{
		"receivers": map[string]interface{}{
			"otlp":     map[string]interface{}{"protocols": map[string]interface{}{"grpc": map[string]interface{}{"endpoint": "127.0.0.1:9319"}}},
			"file_log": map[string]interface{}{"include": []interface{}{"/var/log/*.log"}},
		},
		"processors": map[string]interface{}{
			"batch":             map[string]interface{}{"timeout": "10s"},
			"resourcedetection": map[string]interface{}{},
		},
		"exporters": map[string]interface{}{"otlp_grpc/2": map[string]interface{}{"endpoint": "${env:MW_TARGET}"}},
		"service": map[string]interface{}{"pipelines": map[string]interface{}{
			"logs":    map[string]interface{}{"receivers": []interface{}{"otlp", "file_log"}, "processors": []interface{}{"batch"}, "exporters": []interface{}{"otlp_grpc/2"}},
			"metrics": map[string]interface{}{"receivers": []interface{}{"otlp"}, "processors": []interface{}{"resourcedetection", "batch"}, "exporters": []interface{}{"otlp_grpc/2"}},
		}},
	}
	rules := func(cfg map[string]interface{}) []byte {
		b, err := json.Marshal(map[string]interface{}{"status": true, "config": map[string]interface{}{"nodocker": cfg}})
		require.NoError(t, err)
		return b
	}
	const edits = `[{"op":"set",         "path":["processors","batch","timeout"], "value":"2s"},
 {"op":"remove",      "path":["receivers","file_log"]},
 {"op":"list_remove", "path":["service","pipelines","logs","receivers"], "value":"file_log"},
 {"op":"list_set",    "path":["service","pipelines","metrics","processors"], "value":["batch","resourcedetection"]}]`

	// what the user's edits do to a parsed base
	expect := func(base map[interface{}]interface{}) map[interface{}]interface{} {
		yamlAt(base, "processors", "batch").(map[interface{}]interface{})["timeout"] = "2s"
		delete(yamlAt(base, "receivers").(map[interface{}]interface{}), "file_log")
		yamlAt(base, "service", "pipelines", "logs").(map[interface{}]interface{})["receivers"] = []interface{}{"otlp"}
		yamlAt(base, "service", "pipelines", "metrics").(map[interface{}]interface{})["processors"] = []interface{}{"batch", "resourcedetection"}
		return base
	}
	agent := newFormatTestAgent(t, nil)

	v1 := render(t, agent, rules(baseV1), edits)
	assert.Equal(t, []string{EditApplied, EditApplied, EditApplied, EditApplied}, opResults(v1.Status.Ops))
	final1 := parseRendered(t, v1.Final)
	assert.Equal(t, expect(parseRendered(t, v1.Base)), final1)
	assert.Equal(t, "2s", yamlAt(final1, "processors", "batch", "timeout"))
	assert.Equal(t, []interface{}{"otlp"}, yamlAt(final1, "service", "pipelines", "logs", "receivers"))

	// v2: an integration adds postgres.
	baseV2 := deepCopyConfigValue(baseV1).(map[string]interface{})
	yamlRecv := baseV2["receivers"].(map[string]interface{})
	yamlRecv["postgresql"] = map[string]interface{}{"endpoint": "db:5432", "username": "mw", "password": "pw", "tls": map[string]interface{}{"insecure": true}}
	metrics := baseV2["service"].(map[string]interface{})["pipelines"].(map[string]interface{})["metrics"].(map[string]interface{})
	metrics["receivers"] = []interface{}{"otlp", "postgresql"}

	v2 := render(t, agent, rules(baseV2), edits)
	assert.Equal(t, []string{EditApplied, EditApplied, EditApplied, EditApplied}, opResults(v2.Status.Ops))
	final2 := parseRendered(t, v2.Final)
	assert.Equal(t, expect(parseRendered(t, v2.Base)), final2)
	assert.Equal(t, "db:5432", yamlAt(final2, "receivers", "postgresql", "endpoint"), "postgres kept")
	assert.Nil(t, yamlAt(final2, "receivers", "file_log"), "file_log stays removed")
	assert.Equal(t, "2s", yamlAt(final2, "processors", "batch", "timeout"), "user's 2s kept")
	assert.Equal(t, []interface{}{"otlp", "postgresql"}, yamlAt(final2, "service", "pipelines", "metrics", "receivers"))
	assert.Equal(t, []interface{}{"batch", "resourcedetection"}, yamlAt(final2, "service", "pipelines", "metrics", "processors"))
}

// An edited config that fails to load or validate falls back to the base,
// with the error, and never reports track status.
func TestEditsInvalidEditedConfig(t *testing.T) {
	cases := map[string]struct{ edits, mention string }{
		"pipeline uses an undefined receiver": {
			`[{"op":"list_add","path":["service","pipelines","logs","receivers"],"value":"nosuch"}]`, "nosuch"},
		"unknown component type": {
			`[{"op":"add","path":["receivers","bogus_type"],"value":{}}]`, "bogus_type"},
		"invalid setting": {
			`[{"op":"set","path":["processors","batch","not_a_setting"],"value":1}]`, "not_a_setting"},
	}
	body := readFormatFixture(t, "host.json")
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var trackCalls atomic.Int32
			agent := newFormatTestAgent(t, withTrackCounter(t, &trackCalls))
			r := render(t, agent, body, tc.edits)
			assert.Equal(t, string(r.Base), string(r.Final))
			assert.False(t, r.Status.Applied)
			assert.Contains(t, r.Status.Error, tc.mention)
			assert.Equal(t, []string{EditApplied}, opResults(r.Status.Ops), "per-entry results are kept")
			assert.Zero(t, trackCalls.Load(), "the edited copy must not report track status")
		})
	}
}

// A base that fails validation is an error as today, even with edits that
// would fix it, and track status is reported once.
func TestEditsBaseInvalid(t *testing.T) {
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(readFormatFixture(t, "host.json"), &cfg))
	logs := cfg["config"].(map[string]interface{})["nodocker"].(map[string]interface{})["service"].(map[string]interface{})["pipelines"].(map[string]interface{})["logs"].(map[string]interface{})
	logs["receivers"] = []interface{}{"otlp", "missing"}
	body, err := json.Marshal(cfg)
	require.NoError(t, err)

	var trackCalls atomic.Int32
	agent := newFormatTestAgent(t, withTrackCounter(t, &trackCalls))
	_, buildErr := agent.BuildOtelConfig(body, "nodocker")
	require.ErrorIs(t, buildErr, ErrInvalidConfig)
	require.EqualValues(t, 1, trackCalls.Load())

	r, err := agent.RenderOtelConfig(body, "nodocker",
		[]byte(`[{"op":"list_remove","path":["service","pipelines","logs","receivers"],"value":"missing"}]`))
	assert.Nil(t, r)
	require.ErrorIs(t, err, ErrInvalidConfig)
	assert.Equal(t, buildErr.Error(), err.Error())
	assert.EqualValues(t, 2, trackCalls.Load(), "one track call per render, from the base")
}

// The two branches share no maps or slices: rendering with edits leaves the
// base as it was, and the base's overlays don't leak into the edited copy.
func TestEditsBranchesAreIndependent(t *testing.T) {
	agent := newFormatTestAgent(t, func(c *HostConfig) {
		c.HostTags = "env:prod"
		c.AgentFeatures.LogCollection = false
	})
	body := readFormatFixture(t, "postgres.json") // credential merge adds map[interface{}]interface{} values
	before, err := agent.BuildOtelConfig(body, "nodocker")
	require.NoError(t, err)

	r := render(t, agent, body, `[
	 {"op":"set","path":["receivers","postgresql","databases"],"value":["other"]},
	 {"op":"set","path":["receivers","postgresql","tls","insecure_skip_verify"],"value":true},
	 {"op":"set","path":["processors","batch","timeout"],"value":"2s"},
	 {"op":"list_set","path":["service","pipelines","metrics","processors"],"value":[]},
	 {"op":"list_remove","path":["service","pipelines","metrics","receivers"],"value":"otlp"}]`)
	require.True(t, r.Status.Applied, r.Status.Error)
	assert.Equal(t, string(before), string(r.Base))
	assert.NotEqual(t, string(r.Base), string(r.Final))

	after, err := agent.BuildOtelConfig(body, "nodocker")
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after))

	// The copy taken before the overlays is unaffected by them.
	cfg, err := agent.renderRulesConfig(body, "nodocker")
	require.NoError(t, err)
	copied := deepCopyConfigValue(cfg).(map[string]interface{})
	pristine, err := yaml.Marshal(copied)
	require.NoError(t, err)
	_, err = agent.finishOtelConfig(cfg)
	require.NoError(t, err)
	still, err := yaml.Marshal(copied)
	require.NoError(t, err)
	assert.Equal(t, string(pristine), string(still))
}

func TestDeepCopyConfigValue(t *testing.T) {
	orig := map[string]interface{}{
		"a": map[interface{}]interface{}{"b": []interface{}{map[string]interface{}{"c": 1}}},
		"s": []string{"x"},
		"m": []map[string]interface{}{{"k": "v"}},
	}
	cp := deepCopyConfigValue(orig).(map[string]interface{})
	cp["a"].(map[interface{}]interface{})["b"].([]interface{})[0].(map[string]interface{})["c"] = 2
	cp["s"].([]string)[0] = "y"
	cp["m"].([]map[string]interface{})[0]["k"] = "w"
	cp["new"] = true

	assert.Equal(t, 1, orig["a"].(map[interface{}]interface{})["b"].([]interface{})[0].(map[string]interface{})["c"])
	assert.Equal(t, "x", orig["s"].([]string)[0])
	assert.Equal(t, "v", orig["m"].([]map[string]interface{})[0]["k"])
	assert.NotContains(t, orig, "new")
}

func TestEditsHashIsSha256OfFileBytes(t *testing.T) {
	agent := newFormatTestAgent(t, nil)
	body := readFormatFixture(t, "host.json")
	for _, doc := range []string{`[]`, "[ ]\n", `{"bad":true}`} {
		sum := sha256.Sum256([]byte(doc))
		r := render(t, agent, body, doc)
		assert.Equal(t, hex.EncodeToString(sum[:]), r.Status.EditsHash, "doc %q", doc)
	}
}

func TestEditsStatusJSON(t *testing.T) {
	agent := newFormatTestAgent(t, nil)
	body := readFormatFixture(t, "host.json")

	doc := `[{"op":"set","path":["processors","batch","timeout"],"value":"2s"},{"op":"remove","path":["receivers","nope"]}]`
	r := render(t, agent, body, doc)
	got, err := json.Marshal(r.Status)
	require.NoError(t, err)
	assert.Equal(t, `{"edits_hash":"`+editsHash([]byte(doc))+`","applied":true,"error":"",`+
		`"ops":[{"i":0,"result":"applied"},{"i":1,"result":"skipped","reason":"component receivers/nope not in base"}]}`, string(got))

	doc = `[{"op":"bogus"}]`
	r = render(t, agent, body, doc)
	got, err = json.Marshal(r.Status)
	require.NoError(t, err)
	assert.Equal(t, `{"edits_hash":"`+editsHash([]byte(doc))+`","applied":false,`+
		`"error":"malformed edits: entry 0: unknown op \"bogus\"","ops":[]}`, string(got))

	// The error is capped at 2048 characters.
	doc = `[{"op":"` + strings.Repeat("é", 3000) + `"}]`
	r = render(t, agent, body, doc)
	assert.Len(t, []rune(r.Status.Error), maxEditsErrorLen)
	assert.True(t, strings.HasPrefix(r.Status.Error, "malformed edits: entry 0: unknown op"))
}

// The restrictions run after the edits: with log collection off, a logs
// pipeline added by an edit is still removed, and edits still see the logs
// pipeline the restrictions remove (they apply before them).
func TestEditsAddedLogsPipelineIsRemovedWhenLogCollectionDisabled(t *testing.T) {
	agent := newFormatTestAgent(t, func(c *HostConfig) { c.AgentFeatures.LogCollection = false })
	body := readFormatFixture(t, "host.json")
	r := render(t, agent, body, `[{"op":"add","path":["service","pipelines","logs/app"],"value":{"receivers":["otlp"],"exporters":["otlp_grpc/2"]}},
	 {"op":"list_remove","path":["service","pipelines","logs","receivers"],"value":"file_log"}]`)
	assert.True(t, r.Status.Applied)
	assert.Equal(t, []string{EditApplied, EditApplied}, opResults(r.Status.Ops))
	pipelines := yamlAt(parseRendered(t, r.Final), "service", "pipelines").(map[interface{}]interface{})
	assert.NotContains(t, pipelines, "logs/app")
	assert.NotContains(t, pipelines, "logs")
	assert.Contains(t, pipelines, "metrics")
}

// The same input and edits give the same bytes every time.
func TestEditsRenderIsDeterministic(t *testing.T) {
	agent := newFormatTestAgent(t, func(c *HostConfig) { c.HostTags = "env:prod" })
	body := readFormatFixture(t, "postgres.json")
	doc := `[{"op":"add","path":["receivers","otlp/b"],"value":{"protocols":{"http":{"endpoint":"127.0.0.1:4318"},"grpc":{"endpoint":"127.0.0.1:4317"}}}},
	 {"op":"list_add","path":["service","pipelines","metrics","receivers"],"value":"otlp/b"},
	 {"op":"set","path":["receivers","postgresql","metrics"],"value":{"postgresql.wal.lag":{"enabled":true},"postgresql.deadlocks":{"enabled":true}}}]`
	first := render(t, agent, body, doc)
	require.True(t, first.Status.Applied, first.Status.Error)
	firstStatus, _ := json.Marshal(first.Status)
	for i := 0; i < 20; i++ {
		r := render(t, agent, body, doc)
		require.Equal(t, string(first.Final), string(r.Final))
		require.Equal(t, string(first.Base), string(r.Base))
		status, _ := json.Marshal(r.Status)
		require.Equal(t, string(firstStatus), string(status))
	}
}
