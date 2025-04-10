package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	"go.opentelemetry.io/collector/confmap/provider/envprovider"
	"go.opentelemetry.io/collector/confmap/provider/fileprovider"
	"go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	yaml "gopkg.in/yaml.v2"
)

var (
	ErrRestartAgent  = errors.New("restart agent due to config change")
	ErrInvalidConfig = errors.New("invalid config received from backend")
)

// HostAgent implements Agent interface for Hosts (e.g Linux)
type HostAgent struct {
	HostConfig
	collectorFactories otelcol.Factories
	collectorSettings  otelcol.CollectorSettings
	collector          *otelcol.Collector
	collectorWG        *sync.WaitGroup
	zapCore            zapcore.Core
	logger             *zap.Logger
	httpGetFunc        func(url string) (resp *http.Response, err error)
	Version            string
}

// HostOptions takes in various options for HostAgent
type HostOptions func(h *HostAgent)

// WithHostAgentVersion sets the agent version
func WithHostAgentVersion(v string) HostOptions {
	return func(h *HostAgent) {
		h.Version = v
	}
}

// WithHostAgentLogger sets the logger to be used with agent logs
func WithHostAgentZapCore(zapCore zapcore.Core) HostOptions {
	return func(h *HostAgent) {
		h.zapCore = zapCore
	}
}

// WithHostAgentIsECSEC2 sets whether the agent is running on
// AWS ECS with EC2 infrastructure
func WithHostAgentInfraPlatform(p InfraPlatform) HostOptions {
	return func(h *HostAgent) {
		h.InfraPlatform = p
	}
}

// NewHostAgent returns new agent for Kubernetes with given options.
func NewHostAgent(cfg HostConfig, zapCore zapcore.Core,
	opts ...HostOptions) (*HostAgent, error) {
	var agent HostAgent
	agent.HostConfig = cfg
	agent.httpGetFunc = http.Get

	for _, apply := range opts {
		apply(&agent)
	}

	agent.logger = zap.New(zapCore, zap.AddCaller())

	collectorFactories, err := agent.getFactories()
	if err != nil {
		return nil, err
	}

	agent.collectorFactories = collectorFactories
	agent.collectorSettings = otelcol.CollectorSettings{
		DisableGracefulShutdown: true,
		LoggingOptions: func() []zap.Option {
			// if logfile is specified, then write logs to the file using zapFileCore
			if cfg.Logfile != "" {
				return []zap.Option{
					zap.WrapCore(func(core zapcore.Core) zapcore.Core {
						return zapCore
					}),
				}
			}
			return []zap.Option{}
		}(),

		BuildInfo: component.BuildInfo{
			Command:     "mw-otelcontribcol",
			Description: "Middleware OpenTelemetry Collector Contrib",
			Version:     agent.Version,
		},

		Factories: func() (otelcol.Factories, error) {
			return agent.getFactories()
		},
		ConfigProviderSettings: agent.getConfigProviderSettings(agent.OtelConfigFile),
	}
	agent.collectorWG = &sync.WaitGroup{}

	return &agent, nil
}

var (
	ErrKeyNotFound     = fmt.Errorf("'%s' key not found", Receivers)
	ErrParseReceivers  = fmt.Errorf("failed to parse %s in otel config file", Receivers)
	ErrParseProcessors = fmt.Errorf("failed to parse %s in otel config file", Processors)
	ErrParseService    = fmt.Errorf("failed to parse %s in otel config file", Service)
	ErrParsePipelines  = fmt.Errorf("failed to parse %s in otel config file", Pipelines)
	ErrParseMetrics    = fmt.Errorf("failed to parse %s in otel config file", Metrics)
)

type configType struct {
	Docker     map[string]interface{} `json:"docker"`
	NoDocker   map[string]interface{} `json:"nodocker"`
	Deployment map[string]interface{} `json:"deployment"`
	DaemonSet  map[string]interface{} `json:"daemonset"`
}

// IntegrationType represents the type of the database.
type IntegrationType int

const (
	PostgreSQL IntegrationType = iota
	MongoDB
	MySQL
	MariaDB
	Redis
	Cassandra
	Elasticsearch
	Clickhouse
)

type integrationConfiguration struct {
	// In future this struct can be extended to further accomodate new integrations.
	Path     string `json:"path"`
	Endpoint string `json:"endpoint"`
}

type apiResponseForYAML struct {
	Status              bool                     `json:"status"`
	Config              configType               `json:"config"`
	PgdbConfig          integrationConfiguration `json:"pgdb_config"`
	MongodbConfig       integrationConfiguration `json:"mongodb_config"`
	MysqlConfig         integrationConfiguration `json:"mysql_config"`
	MariaDBConfig       integrationConfiguration `json:"mariadb_config"`
	RedisConfig         integrationConfiguration `json:"redis_config"`
	ElasticsearchConfig integrationConfiguration `json:"elasticsearch_config"`
	CassandraConfig     integrationConfiguration `json:"cassandra_config"`
	ClickhouseConfig    integrationConfiguration `json:"clickhouse_config"`
	Message             string                   `json:"message"`
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

type TrackingMetadata struct {
	HostID        string `json:"host_id"`
	Platform      string `json:"platform"`
	AgentVersion  string `json:"agent_version"`
	InfraPlatform string `json:"infra_platform"`
	Reason        string `json:"reason"`
}
type TrackingPayload struct {
	Status   string           `json:"status"`
	Metadata TrackingMetadata `json:"metadata"`
}

var (
	apiPathForYAML    = "api/v1/agent/ingestion-rules"
	apiPathForRestart = "api/v1/agent/restart-status"
	apiAgentTrack     = "api/v1/agent/tracking"
)

func (d IntegrationType) String() string {
	switch d {
	case PostgreSQL:
		return "postgresql"
	case MongoDB:
		return "mongodb"
	case MySQL:
		return "mysql"
	case MariaDB:
		return "mariadb"
	case Redis:
		return "redis"
	case Cassandra:
		return "cassandra"
	case Elasticsearch:
		return "elasticsearch"
	case Clickhouse:
		return "clickhouse"
	}
	return "unknown"
}

func (c *HostAgent) getConfigProviderSettings(uri string) otelcol.ConfigProviderSettings {
	return otelcol.ConfigProviderSettings{
		ResolverSettings: confmap.ResolverSettings{
			ProviderFactories: []confmap.ProviderFactory{
				fileprovider.NewFactory(),
				yamlprovider.NewFactory(),
				envprovider.NewFactory(),
			},
			URIs: []string{uri},
		},
	}
}
func convertTabsToSpaces(input []byte, tabWidth int) []byte {
	// Find the tab character in the input
	tabChar := byte('\t')

	// Calculate the number of spaces needed to replace each tab
	spaces := bytes.Repeat([]byte(" "), tabWidth)

	// Replace tabs with spaces using bytes.Replace
	output := bytes.Replace(input, []byte{tabChar}, spaces, -1)

	return output
}

func (c *HostAgent) updateConfigWithRestrictions(config map[string]interface{}) (map[string]interface{}, error) {

	receiversData, ok := config[Receivers].(map[string]interface{})
	if !ok {
		return nil, ErrParseReceivers
	}

	serviceData, ok := config[Service].(map[string]interface{})
	if !ok {
		return nil, ErrParseService
	}

	pipelinesData, ok := serviceData[Pipelines].(map[string]interface{})
	if !ok {
		return nil, ErrParsePipelines
	}

	for key, _ := range pipelinesData {
		if !c.HostConfig.AgentFeatures.LogCollection && strings.HasPrefix(key, "logs") {
			delete(pipelinesData, key)
		}

		if !c.HostConfig.AgentFeatures.MetricCollection && strings.HasPrefix(key, "metrics") {
			delete(pipelinesData, key)
		}
	}

	if !c.HostConfig.AgentFeatures.LogCollection {
		delete(receiversData, "filelog")
		delete(receiversData, "windowseventlog")
	}

	if !c.HostConfig.AgentFeatures.MetricCollection {
		delete(receiversData, "hostmetrics")
		delete(receiversData, "windowsperfcounters")
		delete(receiversData, "docker_stats")
		delete(receiversData, "prometheus")
		delete(receiversData, "kubeletstats")
		delete(receiversData, "k8s_cluster")
	}

	return config, nil
}

func (c *HostAgent) updateConfig(config map[string]interface{}, cnf integrationConfiguration) (map[string]interface{}, error) {

	if c.isIPPortFormat(cnf.Endpoint) {
		return config, nil
	}
	// Read the YAML file
	yamlData, err := os.ReadFile(cnf.Path)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Unmarshal the YAML data into a temporary map[string]interface{}
	updatedYamlData := convertTabsToSpaces(yamlData, 2)
	tempMap := make(map[string]interface{})
	err = yaml.Unmarshal(updatedYamlData, &tempMap)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Add the temporary map to the existing "receiver" key
	receiverData, ok := config["receivers"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, ErrKeyNotFound
	}

	for key, value := range tempMap {
		mapValue, mapValueOk := value.(map[interface{}]interface{})
		if mapValueOk {
			oldValue, oldValueOk := receiverData[key]
			if oldValueOk {
				oldMapValue, oldMapValueOk := oldValue.(map[string]interface{})
				if oldMapValueOk {
					for k, v := range mapValue {
						strKey, keyOk := k.(string)
						if keyOk {
							oldMapValue[strKey] = v
						} else {
							c.logger.Info("invalid key type", zap.Any("key type", k))
						}
					}
					receiverData[key] = oldMapValue
				}
			}
		}
	}

	return config, nil
}

func (c *HostAgent) updateConfigFile(configType string) error {
	// _, apiURLForYAML := checkForConfigURLOverrides()

	hostname := getHostname()

	// Call Webhook
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}

	baseURL := u.JoinPath(apiPathForYAML).JoinPath(c.APIKey)
	params := url.Values{}
	params.Add("config", configType)
	params.Add("platform", runtime.GOOS)
	params.Add("host_id", hostname)
	params.Add("host_tags", c.HostTags)
	params.Add("agent_version", c.Version)
	params.Add("infra_platform", fmt.Sprint(c.InfraPlatform))

	collectorRunning := 0
	// Don't need to take lock on the c.collector because it is not deferenced
	if c.collector == nil {
		collectorRunning = 1
	}
	params.Add("col_running", fmt.Sprintf("%d", collectorRunning))

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	url := baseURL.String()
	resp, err := c.httpGetFunc(url)
	if err != nil {
		return fmt.Errorf("failed to call get configuration api for %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get configuration api returned non-200 status: %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal JSON response into ApiResponse struct
	var apiResponse apiResponseForYAML
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal api response: %w", err)
	}

	// Verify API Response
	if !apiResponse.Status {
		return fmt.Errorf("failure status from api response for ingestion rules: %t", apiResponse.Status)
	}

	var apiYAMLConfig map[string]interface{}
	if len(apiResponse.Config.Docker) == 0 && len(apiResponse.Config.NoDocker) == 0 {
		return fmt.Errorf("failed to get valid response, config docker len: %d, config no docker len: %d",
			len(apiResponse.Config.Docker), len(apiResponse.Config.NoDocker))
	}

	apiYAMLConfig = apiResponse.Config.NoDocker
	if configType == "docker" {
		apiYAMLConfig = apiResponse.Config.Docker
	}

	integrationConfigs := map[IntegrationType]integrationConfiguration{
		PostgreSQL:    apiResponse.PgdbConfig,
		MongoDB:       apiResponse.MongodbConfig,
		MySQL:         apiResponse.MysqlConfig,
		MariaDB:       apiResponse.MariaDBConfig,
		Redis:         apiResponse.RedisConfig,
		Elasticsearch: apiResponse.ElasticsearchConfig,
		Cassandra:     apiResponse.CassandraConfig,
		Clickhouse:    apiResponse.ClickhouseConfig,
	}

	for integrationType, integrationConfig := range integrationConfigs {
		if c.checkIntConfigValidity(integrationType, integrationConfig) {
			apiYAMLConfig, err = c.updateConfig(apiYAMLConfig, integrationConfig)
			if err != nil {
				return err
			}
		}
	}

	// Add awsecscontainermetrics receiver dynamically if the agent is running inside ECS + Fargate setup
	if c.InfraPlatform == InfraPlatformECSFargate || c.InfraPlatform == InfraPlatformECSEC2 {

		apiYAMLConfig, err = c.updateConfigForECS(apiYAMLConfig)
		if err != nil {
			return err
		}

	}

	if !c.AgentFeatures.LogCollection || !c.AgentFeatures.MetricCollection {
		apiYAMLConfig, err = c.updateConfigWithRestrictions(apiYAMLConfig)
		if err != nil {
			return err
		}
	}

	// Adding host tags as resource attributes
	if c.HostTags != "" {
		apiYAMLConfig, err = c.updateConfigForHostTags(apiYAMLConfig)
		if err != nil {
			return err
		}
	}

	apiYAMLBytes, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal api data: %w", err)
	}

	// check if the config is valid, otherwise return an error
	factories, _ := c.getFactories()
	cfgProviderSettings := c.getConfigProviderSettings("yaml:" + string(apiYAMLBytes))
	configProvider, err := otelcol.NewConfigProvider(cfgProviderSettings)
	if err != nil {
		return err
	}

	cfg, err := configProvider.Get(context.Background(), factories)
	if err != nil {
		return err
	}

	if err := cfg.Validate(); err != nil {
		trackErr := c.UpdateAgentTrackStatus(err)
		if trackErr != nil {
			c.logger.Error("failed to update agent track status", zap.Error(trackErr))
		}
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	if err := os.WriteFile(c.OtelConfigFile, apiYAMLBytes, 0644); err != nil {
		return fmt.Errorf("failed to write new configuration data to file %s: %w", c.OtelConfigFile, err)
	}

	return nil
}

// GetUpdatedYAMLPath gets the correct otel configuration file
func (c *HostAgent) getOtelConfig() (string, error) {
	configType := "docker"
	dockerSocketPath := strings.Split(c.DockerEndpoint, "//")
	if len(dockerSocketPath) != 2 || !isSocketFn(dockerSocketPath[1]) {
		configType = "nodocker"
	}

	if err := c.updateConfigFile(configType); err != nil {
		return c.OtelConfigFile, err
	}

	return c.OtelConfigFile, nil
}

// isIPPortFormat checks if the given endpoint string is in the format of IPv4:PORT.
// The function also considers localhost as an IP.
// The function returns true if the endpoint matches the regex pattern, false otherwise.
func (c *HostAgent) isIPPortFormat(endpoint string) bool {
	// Regex for IPv4:PORT format, also consider localhost as an IP
	regexPattern := `(?:localhost|((?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?))):(?:[0-9]|[1-9][0-9]{1,4}|[1-5][0-9]{4}|6[0-5][0-5][0-3][0-5])$`

	// Compile the regular expression
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		c.logger.Error("error invalid regexPattern for IP:Port format", zap.Error(err))
		return false
	}

	// Path variable is IP:PORT for these integrations -> Clickhouse, Redpanda
	if regex.MatchString(endpoint) {
		c.logger.Info("found valid endpoint", zap.String("endpoint", endpoint))
		return true
	}
	return false
}

func (c *HostAgent) checkIntConfigValidity(integrationType IntegrationType, cnf integrationConfiguration) bool {
	if cnf.Path != "" {
		// Check if the file exists
		if _, err := os.Stat(cnf.Path); os.IsNotExist(err) {
			c.logger.Warn(fmt.Sprintf("%v config file not found", integrationType), zap.String("path", cnf.Path))
			return false
		}

		return true
	} else if cnf.Endpoint != "" {
		if c.isIPPortFormat(cnf.Endpoint) {
			return true
		}
	}
	return false
}

func (c *HostAgent) callRestartStatusAPI() error {

	// apiURLForRestart, _ := checkForConfigURLOverrides()
	hostname := getHostname()
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}

	baseURL := u.JoinPath(apiPathForRestart)
	baseURL = baseURL.JoinPath(c.APIKey)
	params := url.Values{}
	params.Add("host_id", hostname)
	params.Add("platform", runtime.GOOS)
	params.Add("agent_version", c.Version)
	params.Add("infra_platform", fmt.Sprint(c.InfraPlatform))

	collectorRunning := 0
	// Don't need to take lock on the c.collector because it is not deferenced
	if c.collector == nil {
		collectorRunning = 1
	}
	params.Add("col_running", fmt.Sprintf("%d", collectorRunning))

	// Add Query Parameters to the URL
	baseURL.RawQuery = params.Encode() // Escape Query Parameters

	client := &http.Client{Timeout: 10 * time.Second}
	url := baseURL.String()
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to call restart api for url %s: %w", url, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("restart api returned non-200 status: %d", resp.StatusCode)
	}

	var apiResponse apiResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal restart api response: %w", err)
	}

	if apiResponse.Restart {
		c.logger.Info("fetching updated configuration from backend")
		if _, err := c.getOtelConfig(); err != nil {
			return err
		}

		return ErrRestartAgent
	}

	return err
}

// ListenForConfigChanges listens for configuration changes for the
// agent on the Middleware backend and restarts the agent if configuration
// has changed.
func (c *HostAgent) ListenForConfigChanges(errCh chan<- error,
	stopCh <-chan struct{}) error {

	// First fetch the config
	_, err := c.getOtelConfig()
	if err != nil {
		errCh <- err
	} else {
		errCh <- nil
	}

	restartInterval, err := time.ParseDuration(c.ConfigCheckInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(restartInterval)

	for {
		c.logger.Debug("checking for config change every",
			zap.String("restartInterval", restartInterval.String()))
		select {
		case <-stopCh:
			ticker.Stop()
			return nil
		case <-ticker.C:
			err = c.callRestartStatusAPI()
			errCh <- err
		}
	}
}

func (c *HostAgent) UpdateAgentTrackStatus(reason error) error {
	c.logger.Info("Starting UpdateAgentTrackStatus")
	hostname := getHostname()
	u, err := url.Parse(c.APIURLForConfigCheck)
	if err != nil {
		return err
	}
	baseURL := u.JoinPath(apiAgentTrack)
	baseURL = baseURL.JoinPath(c.APIKey)
	payload := TrackingPayload{
		Status: "validate",
		Metadata: TrackingMetadata{
			HostID:        hostname,
			Platform:      runtime.GOOS,
			AgentVersion:  c.Version,
			InfraPlatform: fmt.Sprint(c.InfraPlatform),
			Reason:        reason.Error(),
		},
	}
	// Marshal payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL.String(), bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	// Add headers
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 10 * time.Second}
	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Agent Track API request failed: %w", err)
	}
	defer resp.Body.Close()
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Agent Track API returned non-200 status code: %d", resp.StatusCode)
	}
	c.logger.Info("Successfully updated agent track status")
	return nil
}

// StartCollector initializes a new OpenTelemetry collector with the configured
// settings and starts it. This function blocks until the collector is stopped
func (c *HostAgent) StartCollector() error {
	if c.collector != nil {
		return nil
	}

	collector, err := otelcol.NewCollector(c.collectorSettings)
	if err != nil {
		return err
	}

	c.collector = collector

	c.collectorWG.Add(1)
	go func() {
		defer c.collectorWG.Done()
		if err := collector.Run(context.Background()); err != nil {
			c.logger.Error("collector server run finished with error",
				zap.Error(err))
			c.collector = nil
		} else {
			c.logger.Info("collector server run finished gracefully")
		}

	}()

	return nil

}

func (c *HostAgent) StopCollector(err error) {
	if c.collector != nil {
		c.logger.Info("stopping telemetry collection", zap.Error(err))
		c.collector.Shutdown()
		c.collectorWG.Wait()
		c.logger.Info("stopped telemetry collection at", zap.Time("time", time.Now()))
		c.collector = nil
	}
}
