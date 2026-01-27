package configupdater

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

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
	// InfraPlatformCycleIO is for Cycle.io platform
	InfraPlatformCycleIO InfraPlatform = 4
	// InfraPlatformEC2 is for AWS EC2 platform
	InfraPlatformEC2 InfraPlatform = 5
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
	case InfraPlatformCycleIO:
		return "cycleio"
	case InfraPlatformEC2:
		return "ec2"
	}
	return "unknown"
}

type configType struct {
	Docker     map[string]interface{} `json:"docker"`
	NoDocker   map[string]interface{} `json:"nodocker"`
	Deployment map[string]interface{} `json:"deployment"`
	DaemonSet  map[string]interface{} `json:"daemonset"`
}

var (
	apiPathForYAML         = "api/v1/agent/ingestion-rules"
	apiPathForRestart      = "api/v1/agent/restart-status"
	apiPathForConfigGroups = "api/v1/agent/public/setting/config-groups" // Apply config class to cluster
)

type apiResponseForYAML struct {
	Status  bool       `json:"status"`
	Config  configType `json:"config"`
	Message string     `json:"message"`
}

type rollout struct {
	Deployment bool `json:"deployment"`
	Daemonset  bool `json:"daemonset"`
}

type apiResponseForRestart struct {
	Status  bool    `json:"status"`
	Restart bool    `json:"restart"`
	Rollout rollout `json:"rollout"`
	Message string  `json:"message"`
}

type Client struct {
	APIURLForConfigCheck string
	APIKey               string
	Target               string
	InfraPlatform        InfraPlatform
}

type AgentFeatures struct {
	MetricCollection    bool
	LogCollection       bool
	SyntheticMonitoring bool
}

// String() implements stringer interface for KubeAgentUpdaterConfig
func (c KubeAgent) String() string {
	var s string
	s += fmt.Sprintf("api-key: %s, ", c.APIKey)
	s += fmt.Sprintf("target: %s, ", c.Target)
	s += fmt.Sprintf("enable-synthetic-monitoring: %t, ", c.EnableSyntheticMonitoring)
	s += fmt.Sprintf("config-check-interval: %s, ", c.ConfigCheckInterval)
	s += fmt.Sprintf("api-url-for-config-check: %s, ", c.APIURLForConfigCheck)
	s += fmt.Sprintf("daemonset-name: %s, ", c.DaemonsetName)
	s += fmt.Sprintf("daemonset-configmap-name: %s, ", c.DaemonsetConfigMapName)
	s += fmt.Sprintf("deployment-name: %s, ", c.DeploymentName)
	s += fmt.Sprintf("deployment-configmap-name: %s, ", c.DeploymentConfigMapName)
	return s
}

type BaseConfig struct {
	APIKey                    string
	Target                    string
	EnableSyntheticMonitoring bool
	ConfigCheckInterval       string
	APIURLForConfigCheck      string
	AgentNamespaceName        string
	DaemonsetName             string
	DaemonsetConfigMapName    string
	DeploymentName            string
	DeploymentConfigMapName   string
	ClusterName               string
	EnableDataDogReceiver     bool
}

// KubeConfig stores configuration for all the host agent
type KubeAgent struct {
	BaseConfig
	clientset           kubernetes.Interface
	configCheckDuration time.Duration
	logger              *zap.Logger
	version             string
	applyConfigOnce     sync.Once
}

func GetAPIURLForConfigCheck(target string) (string, error) {

	// There should at least be two "." in the URL
	parts := strings.Split(target, ".")
	if len(parts) < 3 {
		return "", ErrInvalidTarget
	}

	return strings.TrimSuffix(target, "/"), nil
}

// GetAPIURLForSyntheticMonitoring constructs the WebSocket URL for synthetic monitoring
func GetAPIURLForSyntheticMonitoring(target string) (string, error) {
	// Parse the URL
	parsedURL, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	// Check if the host part of the URL contains more than one '.'
	hostParts := strings.Split(parsedURL.Hostname(), ".")
	if len(hostParts) < 3 {
		return "", ErrInvalidTarget
	}

	// Ensure no trailing slash in the path
	trimmedURL := strings.TrimSuffix(parsedURL.Host, "/")

	// Build the WebSocket URL
	webSocketURL := "wss://" + trimmedURL + "/plsrws/v2"
	return webSocketURL, nil
}
