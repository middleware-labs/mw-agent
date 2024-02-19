package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIURLForConfigCheck(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedResult string
		err            error
	}{
		{
			name:           "URL with both '/' and '.' and without trailing '/'",
			url:            "https://myaccount.middleware.io",
			expectedResult: "https://app.middleware.io",
			err:            nil,
		},
		{
			name:           "URL with trailing '/'",
			url:            "https://myaccount.middleware.io/",
			expectedResult: "https://app.middleware.io",
			err:            nil,
		},
		{
			name:           "URL with only one '.'",
			url:            "https://middleware.io",
			expectedResult: "",
			err:            ErrInvalidTarget,
		},
		{
			name:           "URL with custom domain",
			url:            "https://myaccount.test.mw.io",
			expectedResult: "https://app.test.mw.io",
			err:            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetAPIURLForConfigCheck(test.url)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}

func TestWithKubeAgentMonitorClusterName(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-cluster"
	option := WithKubeAgentMonitorClusterName(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.ClusterName != expected {
		t.Errorf("Expected cluster name %s, got %s", expected, kubeAgentMonitor.ClusterName)
	}
}

func TestWithKubeAgentMonitorDaemonset(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-daemonset"
	option := WithKubeAgentMonitorDaemonset(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.Daemonset != expected {
		t.Errorf("Expected daemonset name %s, got %s", expected, kubeAgentMonitor.Daemonset)
	}
}

func TestWithKubeAgentMonitorDeployment(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-deployment"
	option := WithKubeAgentMonitorDeployment(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.Deployment != expected {
		t.Errorf("Expected deployment name %s, got %s", expected, kubeAgentMonitor.Deployment)
	}
}

func TestWithKubeAgentMonitorAgentNamespace(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-namespace"
	option := WithKubeAgentMonitorAgentNamespace(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.AgentNamespace != expected {
		t.Errorf("Expected agent namespace %s, got %s", expected, kubeAgentMonitor.AgentNamespace)
	}
}

func TestWithKubeAgentMonitorDaemonsetConfigMap(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-daemonset-configmap"
	option := WithKubeAgentMonitorDaemonsetConfigMap(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.DaemonsetConfigMap != expected {
		t.Errorf("Expected daemonset configmap name %s, got %s", expected, kubeAgentMonitor.DaemonsetConfigMap)
	}
}

func TestWithKubeAgentMonitorDeploymentConfigMap(t *testing.T) {
	kubeAgentMonitor := &KubeAgentMonitor{}
	expected := "test-deployment-configmap"
	option := WithKubeAgentMonitorDeploymentConfigMap(expected)
	option(kubeAgentMonitor)
	if kubeAgentMonitor.DeploymentConfigMap != expected {
		t.Errorf("Expected deployment configmap name %s, got %s", expected, kubeAgentMonitor.DeploymentConfigMap)
	}
}
