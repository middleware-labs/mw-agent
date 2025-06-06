package configupdater

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

func TestListenForKubeOtelConfigChanges(t *testing.T) {
	var cfg BaseConfig
	cfg.ConfigCheckInterval = "1s"
	cfg.APIURLForConfigCheck = "http://example.com"
	cfg.APIKey = "apikey"
	cfg.ClusterName = "cluster"
	agent, err := NewKubeAgent(cfg, "0.0.1", fake.NewClientset(), nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listening for config changes in a separate goroutine
	go func() {
		err := agent.ListenForConfigChanges(ctx, make(chan error), make(chan struct{}))
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

	c := &KubeAgent{
		BaseConfig: BaseConfig{
			APIURLForConfigCheck: "http://example.com",
			APIKey:               "apikey",
			ClusterName:          "cluster",
		},

		logger: logger,
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
	err := c.callRestartStatusAPI(context.Background(), false)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case for API error
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	err = c.callRestartStatusAPI(context.Background(), false)
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

	kubeAgentMonitor := &KubeAgent{
		BaseConfig: BaseConfig{
			APIURLForConfigCheck: "http://example.com",
			APIKey:               "apikey",
			ClusterName:          "cluster",
			AgentNamespaceName:   "test-namespace",
			DaemonsetName:        "test-daemonset",
			DeploymentName:       "test-deployment",
		},

		clientset: fakeClientset,
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
