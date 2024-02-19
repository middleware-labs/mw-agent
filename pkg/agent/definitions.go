package agent

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"go.opentelemetry.io/collector/otelcol"
)

// Agent interface provides common methods for different agents
// like host agent (Linux, Windows) & Kubernetes
type Agent interface {
	GetFactories(ctx context.Context) (otelcol.Factories, error)
	GetUpdatedYAMLPath(ctx context.Context) (string, error)
	ListenForConfigChanges(ctx context.Context) error
}

// Otel config components
const (
	Receivers              = "receivers"
	AWSECSContainerMetrics = "awsecscontainermetrics"
	Service                = "service"
	Pipelines              = "pipelines"
	Metrics                = "metrics"
)

var (
	ErrInvalidTarget = fmt.Errorf("invalid target")
)

// InfraPlatform defines the agent's infrastructure platform
type InfraPlatform uint16

var (
	// InfraPlatformInstance is for bare metal or VM platform
	InfraPlatformInstance InfraPlatform = 0
	// InfraPlatformKubernetes is for Kubernetes platform
	InfraPlatformKubernetes InfraPlatform = 1
	// InfraPlatformECSEC2 is for AWS ECS EC2 platform
	InfraPlatformECSEC2 InfraPlatform = 2
	// InfraPlatformECSFargate is for AWS ECS Fargate platform
	InfraPlatformECSFargate InfraPlatform = 3
)

func (p InfraPlatform) String() string {
	switch p {
	case InfraPlatformInstance:
		return "instance"
	case InfraPlatformKubernetes:
		return "kubernetes"
	case InfraPlatformECSEC2:
		return "ecsec2"
	case InfraPlatformECSFargate:
		return "ecsfargate"
	}
	return "unknown"
}

type AgentFeatures struct {
	InfraMonitoring bool
}

// BaseConfig stores general configuration for all agent types
type BaseConfig struct {
	APIKey                    string
	Target                    string
	EnableSyntheticMonitoring bool
	ConfigCheckInterval       string
	FetchAccountOtelConfig    bool
	DockerEndpoint            string
	APIURLForConfigCheck      string
	FluentPort                string
	InfraPlatform             InfraPlatform
	OtelConfigFile            string
	AgentFeatures             AgentFeatures
}

// String() implements stringer interface for BaseConfig
func (c BaseConfig) String() string {
	var s string
	s += fmt.Sprintf("api-key: %s, ", c.APIKey)
	s += fmt.Sprintf("target: %s, ", c.Target)
	s += fmt.Sprintf("enable-synthetic-monitoring: %t, ", c.EnableSyntheticMonitoring)
	s += fmt.Sprintf("config-check-interval: %s, ", c.ConfigCheckInterval)
	s += fmt.Sprintf("docker-endpoint: %s, ", c.DockerEndpoint)
	s += fmt.Sprintf("api-url-for-config-check: %s, ", c.APIURLForConfigCheck)
	s += fmt.Sprintf("infra-platform: %s, ", c.InfraPlatform)
	s += fmt.Sprintf("agent-features: %#v, ", c.AgentFeatures)
	s += fmt.Sprintf("fluent-port: %#v, ", c.FluentPort)
	return s
}

// HostConfig stores configuration for all the host agent
type HostConfig struct {
	BaseConfig

	HostTags    string
	Logfile     string
	LogfileSize int
}

// String() implements stringer interface for HostConfig
func (h HostConfig) String() string {
	s := h.BaseConfig.String()
	s += fmt.Sprintf("host-tags: %s, ", h.HostTags)
	s += fmt.Sprintf("logfile: %s, ", h.Logfile)
	s += fmt.Sprintf("logfile-size: %d", h.LogfileSize)
	return s
}

// KubeConfig stores configuration for all the host agent
type KubeConfig struct {
	BaseConfig
	InsightRefreshDuration string
}

type KubeAgentMonitorConfig struct {
	AgentNamespace      string
	Daemonset           string
	Deployment          string
	DaemonsetConfigMap  string
	DeploymentConfigMap string
}

// WithKubeAgentMonitorClusterName sets the cluster name
func WithKubeAgentMonitorClusterName(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.ClusterName = v
	}
}

// WithKubeAgentMonitorDaemonset sets the daemonset name for the agent
func WithKubeAgentMonitorDaemonset(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.Daemonset = v
	}
}

// WithKubeAgentMonitorDeployment sets the deployment name for the agent
func WithKubeAgentMonitorDeployment(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.Deployment = v
	}
}

// WithKubeAgentMonitorAgentNamespace sets the namespace where the agent is running
func WithKubeAgentMonitorAgentNamespace(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.AgentNamespace = v
	}
}

// WithKubeAgentMonitorDaemonsetConfigMap sets the configmap name for the agent daemonset
func WithKubeAgentMonitorDaemonsetConfigMap(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.DaemonsetConfigMap = v
	}
}

// WithKubeAgentMonitorDeploymentConfigMap sets the configmap name for the agent deployment
func WithKubeAgentMonitorDeploymentConfigMap(v string) KubeAgentMonitorOptions {
	return func(k *KubeAgentMonitor) {
		k.DeploymentConfigMap = v
	}
}

// String() implements stringer interface for KubeConfig
func (k KubeConfig) String() string {
	s := k.BaseConfig.String()
	s += fmt.Sprintf("insight-refresh-duration: %s",
		k.InsightRefreshDuration)

	return s
}

func isSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

var isSocketFn = isSocket

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func GetAPIURLForConfigCheck(target string) (string, error) {
	url := strings.TrimRight(target, "/")

	// There should at least be two "." in the URL
	parts := strings.Split(url, ".")
	if len(parts) < 3 {
		return "", ErrInvalidTarget
	}

	// Find the index of the last "/" and first "."
	firstSlash := strings.LastIndex(url, "/")
	firstDot := strings.Index(url, ".")

	// Check if both "/" and "." exist in the URL
	if firstSlash != -1 && firstDot != -1 {
		// Replace the string between "/" and the first "."
		return url[:firstSlash+1] + "app" + url[firstDot:], nil
	}

	return "", ErrInvalidTarget
}
