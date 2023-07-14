package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
)

func setIsSocketMock(b bool) {
	isSocketFn = func(path string) bool {
		return b
	}
}

func TestGetUpdatedYAMLPath(t *testing.T) {
	agent := NewKubeAgent(
		WithKubeAgentLogger(zap.NewNop()),
		WithKubeAgentDockerEndpoint("unix:///var/run/docker.sock"))

	// Test when docker socket is found
	yamlPath, err := agent.GetUpdatedYAMLPath()
	assert.NoError(t, err)
	assert.Equal(t, "/app/otel-config.yaml", yamlPath)

	// Test when docker socket is not found
	setIsSocketMock(false) // Mock isSocket function to return false
	yamlPath, err = agent.GetUpdatedYAMLPath()
	assert.NoError(t, err)
	assert.Equal(t, "/app/otel-config-nodocker.yaml", yamlPath)

	// Reset the mock
	setIsSocketMock(true)
	yamlPath, err = agent.GetUpdatedYAMLPath()
	assert.NoError(t, err)
	assert.Equal(t, "/app/otel-config.yaml", yamlPath)
}

func TestKubeAgentGetFactories(t *testing.T) {
	agent := NewKubeAgent()

	factories, err := agent.GetFactories(context.Background())
	assert.NoError(t, err)

	// Assert that the returned factories are not nil
	assert.NotNil(t, factories.Extensions)
	assert.NotNil(t, factories.Receivers)
	assert.NotNil(t, factories.Exporters)
	assert.NotNil(t, factories.Processors)

	// check that the returned factories contain the expected factories
	assert.Len(t, factories.Extensions, 0)

	// check if factories contains expected receivers
	assert.Len(t, factories.Receivers, 9)
	assert.Contains(t, factories.Receivers, component.Type("otlp"))
	assert.Contains(t, factories.Receivers, component.Type("fluentforward"))
	assert.Contains(t, factories.Receivers, component.Type("filelog"))
	assert.Contains(t, factories.Receivers, component.Type("docker_stats"))
	assert.Contains(t, factories.Receivers, component.Type("hostmetrics"))
	assert.Contains(t, factories.Receivers, component.Type("k8s_cluster"))
	assert.Contains(t, factories.Receivers, component.Type("k8s_events"))
	assert.Contains(t, factories.Receivers, component.Type("kubeletstats"))
	assert.Contains(t, factories.Receivers, component.Type("prometheus"))

	// check if factories contain expected exporters
	assert.Len(t, factories.Exporters, 3)
	assert.Contains(t, factories.Exporters, component.Type("logging"))
	assert.Contains(t, factories.Exporters, component.Type("otlp"))
	assert.Contains(t, factories.Exporters, component.Type("otlphttp"))

	// check if factories contain expected processors
	assert.Len(t, factories.Processors, 7)
	assert.Contains(t, factories.Processors, component.Type("batch"))
	assert.Contains(t, factories.Processors, component.Type("memory_limiter"))
	assert.Contains(t, factories.Processors, component.Type("filter"))
	assert.Contains(t, factories.Processors, component.Type("resource"))
	assert.Contains(t, factories.Processors, component.Type("resourcedetection"))
	assert.Contains(t, factories.Processors, component.Type("attributes"))
	assert.Contains(t, factories.Processors, component.Type("k8sattributes"))
}
