package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	//	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowseventlogreceiver"
	//	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/windowsperfcountersreceiver"

	"go.uber.org/zap"
	yaml "gopkg.in/yaml.v2"
)

// HostAgent implements Agent interface for Hosts (e.g Linux)
type HostAgent struct {
	HostConfig
	logger      *zap.Logger
	httpGetFunc func(url string) (resp *http.Response, err error)
	Version     string
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
func WithHostAgentLogger(logger *zap.Logger) HostOptions {
	return func(h *HostAgent) {
		h.logger = logger
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
func NewHostAgent(cfg HostConfig, opts ...HostOptions) *HostAgent {
	var agent HostAgent
	agent.HostConfig = cfg
	agent.httpGetFunc = http.Get

	for _, apply := range opts {
		apply(&agent)
	}

	if agent.logger == nil {
		agent.logger, _ = zap.NewProduction()
	}

	return &agent
}

var (
	ErrKeyNotFound    = fmt.Errorf("'%s' key not found", Receivers)
	ErrParseReceivers = fmt.Errorf("failed to parse %s in otel config file", Receivers)
	ErrParseService   = fmt.Errorf("failed to parse %s in otel config file", Service)
	ErrParsePipelines = fmt.Errorf("failed to parse %s in otel config file", Pipelines)
	ErrParseMetrics   = fmt.Errorf("failed to parse %s in otel config file", Metrics)
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

var (
	apiPathForYAML    = "api/v1/agent/ingestion-rules"
	apiPathForRestart = "api/v1/agent/restart-status"
)

func (d IntegrationType) String() string {
	switch d {
	case PostgreSQL:
		return "postgresql"
	case MongoDB:
		return "mongodb"
	case MySQL:
		return "mysql"
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

	apiYAMLBytes, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal api data: %w", err)
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
		return "", err
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

func (c *HostAgent) restartHostAgent() error {
	c.getOtelConfig()
	cmd := exec.Command("kill", "-SIGHUP", fmt.Sprintf("%d", os.Getpid()))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
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
		c.logger.Info("restarting mw-agent")
		if _, err := c.getOtelConfig(); err != nil {
			return fmt.Errorf("error getting updated config: %w", err)
		}

		if err := c.restartHostAgent(); err != nil {
			return fmt.Errorf("error restarting mw-agent: %w", err)
		}
	}

	return err
}

// ListenForConfigChanges listens for configuration changes for the
// agent on the Middleware backend and restarts the agent if configuration
// has changed.
func (c *HostAgent) ListenForConfigChanges(errCh chan<- error,
	stopCh <-chan struct{}) error {

	_, err := c.getOtelConfig()
	if err != nil {
		errCh <- err
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
			return nil
		case <-ticker.C:
			err = c.callRestartStatusAPI()
			errCh <- err
		}
	}
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
