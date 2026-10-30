//go:build linux

package agent

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

const fallbackConfigPath = "../../package-tooling/otel-fallback.yaml"

func newValidateTestAgent(t *testing.T) *HostAgent {
	t.Helper()
	agent := newFormatTestAgent(t, nil)
	t.Setenv("MW_AGENT_GRPC_PORT", "9319")
	t.Setenv("MW_AGENT_HTTP_PORT", "9320")
	t.Setenv("MW_AGENT_FLUENT_PORT", "8006")
	t.Setenv("MW_AGENT_INTERNAL_METRICS_PORT", "8888")
	t.Setenv("MW_AGENT_VERSION", "2.0.0")
	return agent
}

// The startup fallback the packages ship must load with this binary's
// components (a component rename would otherwise leave hosts on the no-op
// pipeline during an outage).
func TestValidateConfigFiles_PackageFallback(t *testing.T) {
	agent := newValidateTestAgent(t)
	require.NoError(t, agent.ValidateConfigFiles(context.Background(), []string{fallbackConfigPath}))
}

// Without a Docker socket the docker_stats receiver fails the collector's
// start, which validate cannot catch: the fallback must not use Docker.
func TestPackageFallbackHasNoDocker(t *testing.T) {
	b, err := os.ReadFile(fallbackConfigPath)
	require.NoError(t, err)
	var cfg struct {
		Receivers  map[string]any `yaml:"receivers"`
		Processors map[string]struct {
			Detectors []string `yaml:"detectors"`
		} `yaml:"processors"`
	}
	require.NoError(t, yaml.Unmarshal(b, &cfg))
	for name := range cfg.Receivers {
		assert.NotContains(t, name, "docker")
	}
	for name, p := range cfg.Processors {
		assert.NotContains(t, p.Detectors, "docker", name)
	}
}

// The fallback is shipped as-is to every release: the agent version comes
// from the binary, and only ${...} references are expanded (a bare $VAR stays
// a literal).
func TestPackageFallbackReferences(t *testing.T) {
	b, err := os.ReadFile(fallbackConfigPath)
	require.NoError(t, err)
	assert.Empty(t, regexp.MustCompile(`\$[A-Za-z_]\w*`).FindAllString(string(b), -1))
	var cfg struct {
		Processors map[string]struct {
			Attributes []struct {
				Key   string `yaml:"key"`
				Value string `yaml:"value"`
			} `yaml:"attributes"`
		} `yaml:"processors"`
	}
	require.NoError(t, yaml.Unmarshal(b, &cfg))
	n := 0
	for name, p := range cfg.Processors {
		for _, a := range p.Attributes {
			if a.Key == "mw.agent.version" {
				assert.Equal(t, "${env:MW_AGENT_VERSION}", a.Value, name)
				n++
			}
		}
	}
	assert.Positive(t, n)
}

func TestValidateConfigFiles_Invalid(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(p, []byte(content), 0o600))
		return p
	}
	for _, tt := range []struct {
		name  string
		files []string
		want  string
	}{
		{"unknown component", []string{write("unknown.yaml", "receivers:\n  bogus: {}\n")}, `unknown type: "bogus"`},
		{"invalid setting", []string{write("interval.yaml", `receivers:
  host_metrics:
    collection_interval: -5s
    scrapers:
      cpu: {}
exporters:
  nop: {}
service:
  pipelines:
    metrics:
      receivers: [host_metrics]
      exporters: [nop]
`)}, "requires positive value"},
		{"merged files, bad reference", []string{fallbackConfigPath, write("ref.yaml", "service:\n  pipelines:\n    metrics:\n      receivers: [nope]\n")}, `references receiver "nope"`},
		{"missing file", []string{filepath.Join(dir, "absent.yaml")}, "no such file"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			agent := newValidateTestAgent(t)
			require.ErrorContains(t, agent.ValidateConfigFiles(context.Background(), tt.files), tt.want)
		})
	}
}
