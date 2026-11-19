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
	"sync"
	"time"

	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/receiver"
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

// knownInfraPlatforms lists every platform ParseInfraPlatform accepts.
var knownInfraPlatforms = []InfraPlatform{
	InfraPlatformInstance,
	InfraPlatformKubernetes,
	InfraPlatformECSEC2,
	InfraPlatformECSFargate,
	InfraPlatformCycleIO,
	InfraPlatformEC2,
}

// ParseInfraPlatform maps a name returned by InfraPlatform.String() back to
// its InfraPlatform. ok is false for an empty or unknown name.
func ParseInfraPlatform(platformName string) (infraPlatform InfraPlatform, ok bool) {
	normalizedName := strings.ToLower(strings.TrimSpace(platformName))
	for _, knownPlatform := range knownInfraPlatforms {
		if knownPlatform.String() == normalizedName {
			return knownPlatform, true
		}
	}
	return InfraPlatformInstance, false
}

type AgentFeatures struct {
	MetricCollection    bool
	LogCollection       bool
	SyntheticMonitoring bool
	OpsAIAutoFix        bool
	ServiceReporting    bool
}

// ExtraReceivers are receiver factories added by an agent's main package.
// The host and kube agents add the obi (eBPF) receiver here on Linux, so
// agents that don't use it (opsai, synthetics) don't import OBI.
var ExtraReceivers []receiver.Factory

// BaseConfig stores general configuration for all agent types
type SyntheticMonitoring struct {
	ApiURL              string
	UnsubscribeEndpoint string
}
type OpsAI struct {
	ApiURL              string
	UnsubscribeEndpoint string
}

type BaseConfig struct {
	APIKey                    string
	Target                    string
	RemoteAgentEnabled        bool
	AgentID                   string
	OpAMPServerURL            string
	OpAMPAPIURL               string
	DaemonsetName             string
	DeploymentName            string
	EnableSyntheticMonitoring bool
	ConfigCheckInterval       string
	FetchAccountOtelConfig    bool
	DockerEndpoint            string
	APIURLForConfigCheck      string
	SyntheticMonitoring
	OpsAI
	GRPCPort              string
	HTTPPort              string
	FluentPort            string
	InfraPlatform         InfraPlatform
	OtelConfigFile        string
	AgentFeatures         AgentFeatures
	SelfProfiling         bool
	ProfilngServerURL     string
	InternalMetricsPort   uint
	EnableDataDogReceiver bool
	// EnableCloudNetworkMonitoring asks the backend for a config with the obi (eBPF) receiver.
	EnableCloudNetworkMonitoring bool
	// EnableEbpfApm asks the backend for an obi receiver that traces the selected services.
	EnableEbpfApm         bool
	EnableInjector        bool
	ServiceReportInterval string
}

// String() implements stringer interface for BaseConfig
func (c BaseConfig) String() string {
	var s string
	s += "api-key: ***, "
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

const (
	// HostIDEnvVar, when set, is used as the host_id the agent sends to the
	// Middleware API instead of resolving it locally. The OpAMP supervisor
	// resolves the host ID once and passes it to the agent through this var.
	HostIDEnvVar = "MW_HOST_ID"
	// InfraPlatformEnvVar, when set to a value of InfraPlatform.String(),
	// skips infra platform detection. Set by the OpAMP supervisor so the
	// supervisor and the agent agree on the platform.
	InfraPlatformEnvVar = "MW_INFRA_PLATFORM"
)

// ec2MetadataBaseURL is the AWS instance metadata service (IMDS) root.
// It is a variable so tests can point it at an httptest server.
var ec2MetadataBaseURL = "http://169.254.169.254/latest"

// getIMDSv2Token retrieves an IMDSv2 token for AWS EC2 metadata service
func getIMDSv2Token() (string, error) {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	tokenURL := ec2MetadataBaseURL + "/api/token"
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

	metadataURL := ec2MetadataBaseURL + "/meta-data/" + metadataPath
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

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

// getEC2Hostname returns the IMDS "hostname" (the private DNS name, such as
// ip-172-31-23-233.eu-north-1.compute.internal). It is the value the OTel
// contrib ec2 resourcedetection detector reports as host.id, so the agent's
// host_id matches the host.id on its telemetry.
func getEC2Hostname() (string, error) {
	ec2Hostname, err := getEC2Metadata("hostname")
	if err != nil {
		return "", err
	}
	ec2Hostname = strings.TrimSpace(ec2Hostname)
	if ec2Hostname == "" {
		return "", fmt.Errorf("EC2 metadata returned an empty hostname")
	}
	return ec2Hostname, nil
}

// cachedEC2Hostname holds the first IMDS hostname read successfully. The
// private DNS name does not change while the instance runs, and the agent asks
// for it on every ingestion-rules and restart-status call. Without the cache a
// single IMDS timeout would send that call with os.Hostname() instead, and the
// config returned for that name would drop the host-level integrations.
var cachedEC2Hostname struct {
	mu       sync.Mutex
	hostname string
}

// resolveEC2Hostname returns the cached IMDS hostname, or reads and caches it.
// A failed read is not cached, so a later call tries IMDS again.
func resolveEC2Hostname() (string, error) {
	cachedEC2Hostname.mu.Lock()
	defer cachedEC2Hostname.mu.Unlock()

	if cachedEC2Hostname.hostname != "" {
		return cachedEC2Hostname.hostname, nil
	}

	ec2Hostname, err := getEC2Hostname()
	if err != nil {
		return "", err
	}
	cachedEC2Hostname.hostname = ec2Hostname
	return ec2Hostname, nil
}

// GetHostnameForPlatform returns the host ID the agent sends to the
// Middleware API. MW_HOST_ID wins when set; on EC2 it is the IMDS hostname,
// read once and reused, falling back to os.Hostname() until IMDS answers;
// elsewhere it is os.Hostname().
func GetHostnameForPlatform(infraPlatform InfraPlatform) string {
	if hostIDOverride := strings.TrimSpace(os.Getenv(HostIDEnvVar)); hostIDOverride != "" {
		return hostIDOverride
	}

	if infraPlatform == InfraPlatformEC2 {
		ec2Hostname, err := resolveEC2Hostname()
		if err != nil {
			return getHostname()
		}
		return ec2Hostname
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
