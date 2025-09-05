package agent

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
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
	Processors             = "processors"
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

type AgentFeatures struct {
	MetricCollection    bool
	LogCollection       bool
	SyntheticMonitoring bool
	OpsAIAutoFix        bool
}

// BaseConfig stores general configuration for all agent types
type BaseConfig struct {
	APIKey                       string
	Target                       string
	EnableSyntheticMonitoring    bool
	ConfigCheckInterval          string
	FetchAccountOtelConfig       bool
	DockerEndpoint               string
	APIURLForConfigCheck         string
	APIURLForSyntheticMonitoring string
	GRPCPort                     string
	HTTPPort                     string
	FluentPort                   string
	InfraPlatform                InfraPlatform
	OtelConfigFile               string
	AgentFeatures                AgentFeatures
	SelfProfiling                bool
	ProfilngServerURL            string
	InternalMetricsPort          uint
	EnableDataDogReceiver        bool
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

	HostTags     string
	Logfile      string
	LogfileSize  int
	LoggingLevel string
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
}

type KubeAgentMonitorConfig struct {
	AgentNamespace      string
	Daemonset           string
	Deployment          string
	DaemonsetConfigMap  string
	DeploymentConfigMap string
}

// WithKubeAgentMonitorVersion sets the agent version
func WithKubeAgentMonitorVersion(v string) KubeAgentMonitorOptions {
	return func(h *KubeAgentMonitor) {
		h.Version = v
	}
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

// getIMDSv2Token retrieves an IMDSv2 token for AWS EC2 metadata service
func getIMDSv2Token() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	tokenURL := "http://169.254.169.254/latest/api/token"
	req, err := http.NewRequest("PUT", tokenURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get IMDSv2 token, status: %d", tokenResp.StatusCode)
	}

	tokenBytes, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", err
	}

	return string(tokenBytes), nil
}

// getEC2Metadata retrieves metadata from AWS EC2 metadata service using IMDSv2
func getEC2Metadata(metadataPath string) (string, error) {
	token, err := getIMDSv2Token()
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	metadataURL := "http://169.254.169.254/latest/meta-data/" + metadataPath
	req, err := http.NewRequest("GET", metadataURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get EC2 metadata from %s, status: %d", metadataPath, resp.StatusCode)
	}

	metadataBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(metadataBytes), nil
}

// isEC2Instance checks if the current instance is running on AWS EC2
// by attempting to contact the AWS Instance Metadata Service (IMDS) v2
func IsEC2Instance() bool {
	_, err := getEC2Metadata("instance-id")
	return err == nil
}

// getEC2Hostname retrieves the internal hostname/FQDN from AWS EC2 metadata service
func getEC2Hostname() (string, error) {
	return getEC2Metadata("local-hostname")
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

// GetHostnameForPlatform returns hostname based on the infra platform
func GetHostnameForPlatform(infraPlatform InfraPlatform) string {
	// Get EC2 full hostname when running on EC2, otherwise use system hostname
	if infraPlatform == InfraPlatformEC2 {
		hostname, err := getEC2Hostname()
		if err != nil {
			// Fall back to system hostname if EC2 hostname retrieval fails
			return getHostname()
		}
		return hostname
	}
	return getHostname()
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
	fmt.Println("webSocketURL", webSocketURL)
	return webSocketURL, nil
}

func HasValidTags(tags string) error {
	if tags == "" {
		return nil
	}
	pairs := strings.Split(tags, ",")
	for _, pair := range pairs {
		keyValue := strings.Split(pair, ":")
		if len(keyValue) != 2 {
			return fmt.Errorf("invalid tag format: %s", pair)
		}
	}
	return nil
}

type Profiler struct {
	Logger        *zap.Logger
	ServerAddress string
}

func NewProfiler(logger *zap.Logger, serverAddress string) *Profiler {
	return &Profiler{
		Logger:        logger,
		ServerAddress: serverAddress,
	}
}

func (p *Profiler) StartProfiling(appName string, target string, tags string) {
	parsedURL, err := url.Parse(target)

	if err != nil {
		p.Logger.Error("PROFILER: Invalid URL - MW_TARGET")
		return
	}

	hostParts := strings.Split(parsedURL.Hostname(), ".")

	if len(hostParts) <= 1 {
		p.Logger.Error("PROFILER: Subdomain doesn't exist - MW_TARGET")
		return
	}

	p.Logger.Info("PROFILER: TenantID-" + hostParts[0])

	config := pyroscope.Config{
		ApplicationName: appName,
		ServerAddress:   p.ServerAddress,
		TenantID:        hostParts[0],
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileAllocSpace,
		},
	}

	if len(tags) > 0 {
		config.Tags = map[string]string{tags: tags}
	}

	_, err = pyroscope.Start(config)

	if err != nil {
		p.Logger.Error("PROFILER: Couldn't run profiler on mw-agent", zap.Error(err))
	}

	p.Logger.Info("PROFILER: Running on mw-agent")
}
