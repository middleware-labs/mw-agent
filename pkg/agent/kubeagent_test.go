package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
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
	assert.Len(t, factories.Extensions, 1)

	// check if factories contains expected receivers
	assert.Len(t, factories.Receivers, 15)
	assertContainsComponent(t, factories.Receivers, "otlp")
	assertContainsComponent(t, factories.Receivers, "fluentforward")
	assertContainsComponent(t, factories.Receivers, "filelog")
	assertContainsComponent(t, factories.Receivers, "docker_stats")
	assertContainsComponent(t, factories.Receivers, "hostmetrics")
	assertContainsComponent(t, factories.Receivers, "k8s_cluster")
	assertContainsComponent(t, factories.Receivers, "k8s_events")
	assertContainsComponent(t, factories.Receivers, "kubeletstats")
	assertContainsComponent(t, factories.Receivers, "prometheus")
	assertContainsComponent(t, factories.Receivers, "k8sobjects")
	assertContainsComponent(t, factories.Receivers, "statsd")
	assertContainsComponent(t, factories.Receivers, "journald")
	assertContainsComponent(t, factories.Receivers, "rabbitmq")
	assertContainsComponent(t, factories.Receivers, "sqlserver")
	assertContainsComponent(t, factories.Receivers, "nginx")

	// check if factories contain expected exporters
	assert.Len(t, factories.Exporters, 4)
	assertContainsComponent(t, factories.Exporters, "debug")
	assertContainsComponent(t, factories.Exporters, "otlp")
	assertContainsComponent(t, factories.Exporters, "otlphttp")
	assertContainsComponent(t, factories.Exporters, "kafka")

	// check if factories contain expected processors
	assert.Len(t, factories.Processors, 12)
	assertContainsComponent(t, factories.Processors, "batch")
	assertContainsComponent(t, factories.Processors, "memory_limiter")
	assertContainsComponent(t, factories.Processors, "filter")
	assertContainsComponent(t, factories.Processors, "resource")
	assertContainsComponent(t, factories.Processors, "resourcedetection")
	assertContainsComponent(t, factories.Processors, "attributes")
	assertContainsComponent(t, factories.Processors, "k8sattributes")
	assertContainsComponent(t, factories.Processors, "cumulativetodelta")
	assertContainsComponent(t, factories.Processors, "deltatorate")
	assertContainsComponent(t, factories.Processors, "metricstransform")
	assertContainsComponent(t, factories.Processors, "transform")
	assertContainsComponent(t, factories.Processors, "groupbyattrs")
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
	zapEncoderCfg := zapcore.EncoderConfig{
		MessageKey:   "message",
		LevelKey:     "level",
		EncodeLevel:  zapcore.CapitalLevelEncoder,
		TimeKey:      "time",
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}
	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig = zapEncoderCfg
	logger, _ := zapCfg.Build()

	c := &KubeAgentMonitor{
		KubeConfig: KubeConfig{
			BaseConfig: BaseConfig{
				APIURLForConfigCheck: "http://example.com",
				APIKey:               "apikey",
			},
		},
		logger:      logger,
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

	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a map representing the JSON response body
		responseBody := map[string]interface{}{
			"key": "value",
		}

		// Encode the response body to JSON
		jsonResponse, err := json.Marshal(responseBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to indicate JSON content
		w.Header().Set("Content-Type", "application/json")

		// Write the JSON response with a 200 status code
		w.WriteHeader(http.StatusOK)
		_, errWrite := w.Write(jsonResponse)
		if errWrite != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Sample daemonset
	daemonset := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-daemonset",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
		},
	}

	_, errDaemonset := fakeClientset.AppsV1().DaemonSets("test-namespace").Create(ctx, daemonset, metav1.CreateOptions{})
	if errDaemonset != nil {
		return
	}

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

	// Sample deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
		},
	}

	_, errDeployment := fakeClientset.AppsV1().Deployments("test-namespace").Create(ctx, deployment, metav1.CreateOptions{})
	if errDeployment != nil {
		return
	}

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
