package configupdater

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
)

func TestComponentTypeString(t *testing.T) {
	tests := []struct {
		name string
		ct   ComponentType
		want string
	}{
		{"Deployment", Deployment, "deployment"},
		{"DaemonSet", DaemonSet, "daemonset"},
		{"Unknown", ComponentType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.ct.String())
		})
	}
}

func TestNewKubeAgent(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger, _ := zap.NewProduction()

	tests := []struct {
		name    string
		cfg     BaseConfig
		opts    []KubeAgentOptions
		wantErr bool
	}{
		{
			"Valid Config",
			BaseConfig{ConfigCheckInterval: "10s"},
			[]KubeAgentOptions{
				WithKubeAgentVersion("1.0.0"),
				WithKubeAgentLogger(logger),
				WithKubeAgentClientset(fakeClient),
				WithKubeAgentHTTPGetFunc(func(string) (*http.Response, error) {
					return &http.Response{StatusCode: http.StatusOK}, nil
				}),
			},
			false,
		},
		{
			"Invalid Config Interval",
			BaseConfig{ConfigCheckInterval: "invalid"},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewKubeAgent(tt.cfg, tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, agent)
			}
		})
	}
}

func TestUpdateConfigMap(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()
	logger, _ := zap.NewProduction()
	ctx := context.TODO()

	tests := []struct {
		name     string
		compType ComponentType
		opts     []KubeAgentOptions
		response apiResponseForYAML
		wantErr  bool
	}{
		{
			name:     "Update ConfigMap",
			compType: Deployment,
			opts: []KubeAgentOptions{
				WithKubeAgentVersion("1.0.0"),
				WithKubeAgentLogger(logger),
				WithKubeAgentClientset(fakeClient),
				WithKubeAgentHTTPGetFunc(func(string) (*http.Response, error) {
					return &http.Response{
						Body:       io.NopCloser(strings.NewReader(`{"status": true, "config": {"deployment": {"test": "test"}}}`)),
						StatusCode: http.StatusOK,
					}, nil
				}),
			},
			response: apiResponseForYAML{
				Status: true,
				Config: configType{
					Deployment: map[string]interface{}{"test": "test"},
				},
			},
			wantErr: false,
		},
		{
			name:     "Server Error",
			compType: ComponentType(99),
			opts: []KubeAgentOptions{
				WithKubeAgentVersion("1.0.0"),
				WithKubeAgentLogger(logger),
				WithKubeAgentClientset(fakeClient),
				WithKubeAgentHTTPGetFunc(func(string) (*http.Response, error) {
					return &http.Response{
						Body:       http.NoBody,
						StatusCode: http.StatusInternalServerError,
					}, nil
				}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewKubeAgent(BaseConfig{ConfigCheckInterval: "10s"}, tt.opts...)
			if err != nil {
				t.Fatal(err)
			}
			err = agent.UpdateConfigMap(ctx, tt.compType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRolloutRestart(t *testing.T) {
	//fakeClient := fake.NewSimpleClientset()
	fakeClient := fake.NewClientset(
		&appsv1.Deployment{
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
		},
		&appsv1.DaemonSet{
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
		},
	)

	ctx := context.TODO()

	tests := []struct {
		name       string
		compType   ComponentType
		baseConfig BaseConfig
		wantErr    bool
	}{
		{
			name:     "Rollout Restart",
			compType: Deployment,
			baseConfig: BaseConfig{
				ConfigCheckInterval: "10s",
				AgentNamespaceName:  "test-namespace",
				DeploymentName:      "test-deployment",
				DaemonsetName:       "test-daemonset",
			},
			wantErr: false,
		},
		{
			name:     "Rollout Restart",
			compType: DaemonSet,
			baseConfig: BaseConfig{
				ConfigCheckInterval: "10s",
				AgentNamespaceName:  "test-namespace",
				DeploymentName:      "test-deployment",
				DaemonsetName:       "test-daemonset",
			},
			wantErr: false,
		},
		{
			name:     "Rollout Restart",
			compType: Deployment,
			baseConfig: BaseConfig{
				ConfigCheckInterval: "10s",
				AgentNamespaceName:  "test-namespace",
				DeploymentName:      "test-deployment-not-exist",
				DaemonsetName:       "test-daemonset",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent, err := NewKubeAgent(tt.baseConfig, WithKubeAgentVersion("1.0.0"),
				WithKubeAgentClientset(fakeClient))
			assert.NoError(t, err)
			err = agent.rolloutRestart(ctx, tt.compType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
