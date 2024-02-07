package agent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

func TestKubeAgentGetFactories(t *testing.T) {
	agent := NewKubeAgent(KubeConfig{})

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

func TestListenForKubeOtelConfigChanges(t *testing.T) {
	cfg := KubeConfig{}
	cfg.ConfigCheckInterval = "1s"
	agent := NewKubeAgentMonitor(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listening for config changes in a separate goroutine
	go func() {
		err := agent.ListenForKubeOtelConfigChanges(ctx)
		assert.NoError(t, err)
	}()

	// Wait for a short period to allow the goroutine to start
	time.Sleep(100 * time.Millisecond)

	// Manually cancel the context to stop listening for config changes
	cancel()

	// Wait for a short period to allow the goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Assert that the goroutine has stopped and no error occurred
	assert.True(t, true)
}

func TestCallRestartStatusAPI(t *testing.T) {
	// Create a KubeAgentMonitor instance with mocked dependencies
	c := &KubeAgentMonitor{
		KubeConfig: KubeConfig{
			BaseConfig: BaseConfig{
				APIURLForConfigCheck: "http://example.com",
				APIKey:               "apikey",
			},
		},

		ClusterName: "cluster",
	}

	// Mock the HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/restart" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Override the API URL with the mock server's URL
	c.APIURLForConfigCheck = server.URL

	// Test case for successful API call
	err := c.callRestartStatusAPI(context.Background())
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case for API error
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	err = c.callRestartStatusAPI(context.Background())
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

func TestRolloutRestart(t *testing.T) {

	// Create a fake clientset for testing
	fakeClientset := fake.NewSimpleClientset()

	// Initialize your struct with the fake clientset
	kubeAgentMonitor := &KubeAgentMonitor{
		Clientset: fakeClientset,
		KubeAgentMonitorConfig: KubeAgentMonitorConfig{
			AgentNamespace: "test-namespace",
			Daemonset:      "test-daemonset",
			Deployment:     "test-deployment",
		},
	}

	// Mocking the DaemonSet update
	fakeClientset.PrependReactor("update", "daemonsets", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if updateAction, ok := action.(kubetesting.UpdateActionImpl); ok {
			obj := updateAction.GetObject()
			return true, obj, nil
		}
		return false, nil, nil
	})

	// Call rolloutRestart for DaemonSet
	err := kubeAgentMonitor.rolloutRestart(context.Background(), DaemonSet)
	assert.NoError(t, err)

	// Mocking the DaemonSet update
	fakeClientset.PrependReactor("update", "deployments", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if updateAction, ok := action.(kubetesting.UpdateActionImpl); ok {
			obj := updateAction.GetObject()
			return true, obj, nil
		}
		return false, nil, nil
	})

	// Call rolloutRestart for DaemonSet
	err = kubeAgentMonitor.rolloutRestart(context.Background(), Deployment)
	assert.NoError(t, err)
}
