package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/extension/healthcheckextension"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/attributesprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourcedetectionprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/resourceprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/dockerstatsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/filelogreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/fluentforwardreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/mongodbreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/postgresqlreceiver"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/prometheusreceiver"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/loggingexporter"
	"go.opentelemetry.io/collector/exporter/otlpexporter"
	"go.opentelemetry.io/collector/exporter/otlphttpexporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/otelcol"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/batchprocessor"
	"go.opentelemetry.io/collector/processor/memorylimiterprocessor"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

// HostAgent implements Agent interface for Hosts (e.g Linux)
type HostAgent struct {
	apiKey string
	target string

	enableSyntheticMonitoring bool
	configCheckInterval       string

	apiURLForConfigCheck string

	logger         *zap.Logger
	dockerEndpoint string
	hostTags       string
}

// HostOptions takes in various options for HostAgent
type HostOptions func(h *HostAgent)

// WithHostAgentApiKey sets api key for interacting with
// the Middleware backend
func WithHostAgentApiKey(key string) HostOptions {
	return func(h *HostAgent) {
		h.apiKey = key
	}
}

// WithHostAgentTarget sets target URL for sending insights
// to the Middlware backend.
func WithHostAgentTarget(t string) HostOptions {
	return func(h *HostAgent) {
		h.target = t
	}
}

// WithHostAgentEnableSyntheticMonitoring enables synthetic
// monitoring to be performed from the agent.
func WithHostAgentEnableSyntheticMonitoring(e bool) HostOptions {
	return func(h *HostAgent) {
		h.enableSyntheticMonitoring = e
	}
}

// WithHostAgentConfigCheckInterval sets the duration for checking with
// the Middleware backend for configuration update.
func WithHostAgentConfigCheckInterval(c string) HostOptions {
	return func(h *HostAgent) {
		h.configCheckInterval = c
	}
}

// WithHostAgentApiURLForConfigCheck sets the URL for the periodic
// configuration check.
func WithHostAgentApiURLForConfigCheck(u string) HostOptions {
	return func(h *HostAgent) {
		h.apiURLForConfigCheck = u
	}
}

// WithHostAgentLogger sets the logger to be used with agent logs
func WithHostAgentLogger(logger *zap.Logger) HostOptions {
	return func(h *HostAgent) {
		h.logger = logger
	}
}

// WithHostAgentDockerEndpoint sets the endpoint for docker so that
// the agent can figure out if it needs to send docker logs & metrics.
func WithHostAgentDockerEndpoint(endpoint string) HostOptions {
	return func(h *HostAgent) {
		h.dockerEndpoint = endpoint
	}
}

// WithHostAgentHostTags sets the tag for particular host
func WithHostAgentHostTags(tags string) HostOptions {
	return func(h *HostAgent) {
		h.hostTags = tags
	}
}

// NewHostAgent returns new agent for Kubernetes with given options.
func NewHostAgent(opts ...HostOptions) *HostAgent {
	var cfg HostAgent
	for _, apply := range opts {
		apply(&cfg)
	}

	if cfg.logger == nil {
		cfg.logger, _ = zap.NewProduction()
	}

	return &cfg
}

var (
	ErrRestartStatusAPINotOK = errors.New("received error code from the server")
	ErrReceiverKeyNotFound   = errors.New("'receivers' key not found")
	ErrInvalidResponse       = errors.New("invalid response from ingestion rules api")
	ErrInvalidHostTags       = errors.New("invalid host tags, kindly check the format")
)

type configType struct {
	Docker   map[string]interface{} `json:"docker"`
	NoDocker map[string]interface{} `json:"nodocker"`
}

type pgdbConfiguration struct {
	Path string `json:"path"`
}

type mongodbConfiguration struct {
	Path string `json:"path"`
}

type apiResponseForYAML struct {
	Status        bool                 `json:"status"`
	Config        configType           `json:"config"`
	PgdbConfig    pgdbConfiguration    `json:"pgdb_config"`
	MongodbConfig mongodbConfiguration `json:"mongodb_config"`
	Message       string               `json:"message"`
}

type apiResponseForRestart struct {
	Status  bool   `json:"status"`
	Restart bool   `json:"restart"`
	Message string `json:"message"`
}

var (
	apiPathForYAML    = "api/v1/agent/ingestion-rules"
	apiPathForRestart = "api/v1/agent/restart-status"
)

const (
	yamlFile         = "configyamls/all/otel-config.yaml"
	yamlFileNoDocker = "configyamls/nodocker/otel-config.yaml"
)

func (c *HostAgent) updatepgdbConfig(config map[string]interface{},
	pgdbConfig pgdbConfiguration) (map[string]interface{}, error) {
	return c.updateConfig(config, pgdbConfig.Path)
}

func (c *HostAgent) updateMongodbConfig(config map[string]interface{},
	mongodbConfig mongodbConfiguration) (map[string]interface{}, error) {
	return c.updateConfig(config, mongodbConfig.Path)
}

func (c *HostAgent) updateConfig(config map[string]interface{}, path string) (map[string]interface{}, error) {

	// Read the YAML file
	yamlData, err := ioutil.ReadFile(path)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Unmarshal the YAML data into a temporary map[string]interface{}
	tempMap := make(map[string]interface{})
	err = yaml.Unmarshal(yamlData, &tempMap)
	if err != nil {
		return map[string]interface{}{}, err
	}

	// Add the temporary map to the existing "receiver" key
	receiverData, ok := config["receivers"].(map[string]interface{})
	if !ok {
		return map[string]interface{}{}, ErrReceiverKeyNotFound
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

func (c *HostAgent) updateYAML(configType, yamlPath string) error {
	// _, apiURLForYAML := checkForConfigURLOverrides()

	hostname := getHostname()

	// Call Webhook
	u, err := url.Parse(c.apiURLForConfigCheck)
	if err != nil {
		return err
	}

	baseUrl := u.JoinPath(apiPathForYAML).JoinPath(c.apiKey)
	params := url.Values{}
	params.Add("config", configType)
	params.Add("platform", runtime.GOOS)
	params.Add("host_id", hostname)
	// Add Query Parameters to the URL
	baseUrl.RawQuery = params.Encode() // Escape Query Parameters

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		c.logger.Error("failed to call get configuration api", zap.Error(err))
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("failed to call get configuration api", zap.Int("statuscode", resp.StatusCode))
		return ErrRestartStatusAPINotOK
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to reas response body", zap.Error(err))
		return err
	}

	// Unmarshal JSON response into ApiResponse struct
	var apiResponse apiResponseForYAML
	// fmt.Println("body: ", string(body))
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		c.logger.Error("failed to unmarshal api response", zap.Error(err))
		return err
	}

	// Verify API Response
	if !apiResponse.Status {
		c.logger.Error("failure status from api response for ingestion rules", zap.Bool("status", apiResponse.Status))
		return ErrInvalidResponse
	}

	var apiYAMLConfig map[string]interface{}
	if len(apiResponse.Config.Docker) == 0 && len(apiResponse.Config.NoDocker) == 0 {
		c.logger.Error("failed to get valid response",
			zap.Int("config docker len", len(apiResponse.Config.Docker)),
			zap.Int("config no docker len", len(apiResponse.Config.NoDocker)))
		return ErrInvalidResponse
	} else {
		if configType == "docker" {
			apiYAMLConfig = apiResponse.Config.Docker
		} else {
			apiYAMLConfig = apiResponse.Config.NoDocker
		}
	}

	pgdbConfig := apiResponse.PgdbConfig
	if pgdbConfig.Path != "" {
		apiYAMLConfig, err = c.updatepgdbConfig(apiYAMLConfig, pgdbConfig)
		if err != nil {
			return err
		}
	}

	mongodbConfig := apiResponse.MongodbConfig
	if mongodbConfig.Path != "" {
		apiYAMLConfig, err = c.updateMongodbConfig(apiYAMLConfig, mongodbConfig)
		if err != nil {
			return err
		}
	}

	apiYAMLBytes, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		c.logger.Error("failed to marshal api data", zap.Error(err))
		return err
	}

	if err := os.WriteFile(yamlPath, apiYAMLBytes, 0644); err != nil {
		c.logger.Error("failed to write new configuration data to file", zap.Error(err))
		return err
	}

	return nil
}

// GetUpdatedYAMLPath gets the correct otel configuration file.
func (c *HostAgent) GetUpdatedYAMLPath() (string, error) {
	configType := "docker"
	yamlPath := yamlFile
	dockerSocketPath := strings.Split(c.dockerEndpoint, "//")
	if len(dockerSocketPath) != 2 || !isSocketFn(dockerSocketPath[1]) {
		configType = "nodocker"
		yamlPath = yamlFileNoDocker
	}

	if err := c.updateYAML(configType, yamlPath); err != nil {
		return yamlPath, err
	}

	return yamlPath, nil
}

func restartHostAgent() error {
	//GetUpdatedYAMLPath()
	cmd := exec.Command("kill", "-SIGHUP", fmt.Sprintf("%d", os.Getpid()))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (c *HostAgent) callRestartStatusAPI() error {

	// fmt.Println("Starting recursive restart check......")
	// apiURLForRestart, _ := checkForConfigURLOverrides()
	hostname := getHostname()
	u, err := url.Parse(c.apiURLForConfigCheck)
	if err != nil {
		return err
	}

	baseUrl := u.JoinPath(apiPathForRestart)
	baseUrl = baseUrl.JoinPath(c.apiKey)
	params := url.Values{}
	params.Add("host_id", hostname)
	params.Add("platform", runtime.GOOS)

	// Add Query Parameters to the URL
	baseUrl.RawQuery = params.Encode() // Escape Query Parameters

	resp, err := http.Get(baseUrl.String())
	if err != nil {
		c.logger.Error("failed to call Restart-API", zap.String("url", baseUrl.String()), zap.Error(err))
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("failed to call Restart-API", zap.Int("code", resp.StatusCode))
		return ErrRestartStatusAPINotOK
	}

	var apiResponse apiResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		c.logger.Error("failed unmarshal Restart-API response", zap.Error(err))
		return err
	}

	if apiResponse.Restart {
		c.logger.Info("restarting mw-agent")
		if err := restartHostAgent(); err != nil {
			c.logger.Error("error restarting mw-agent", zap.Error(err))
			return err
		}
	}

	return err
}

// ListenForConfigChanges listens for configuration changes for the
// agent on the Middleware backend and restarts the agent if configuration
// has changed.
func (c *HostAgent) ListenForConfigChanges(ctx context.Context) error {

	restartInterval, err := time.ParseDuration(c.configCheckInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(restartInterval)

	go func() {
		for {
			c.logger.Info("check for config changes after", zap.Duration("restartInterval", restartInterval))
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.callRestartStatusAPI()
			}
		}
	}()

	return nil
}

// GetFactories get otel factories for HostAgent
func (c *HostAgent) GetFactories(ctx context.Context) (otelcol.Factories, error) {
	var err error
	factories := otelcol.Factories{}
	factories.Extensions, err = extension.MakeFactoryMap(
		healthcheckextension.NewFactory(),
	// frontend.NewAuthFactory(),
	)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Receivers, err = receiver.MakeFactoryMap([]receiver.Factory{
		otlpreceiver.NewFactory(),
		fluentforwardreceiver.NewFactory(),
		filelogreceiver.NewFactory(),
		dockerstatsreceiver.NewFactory(),
		hostmetricsreceiver.NewFactory(),
		prometheusreceiver.NewFactory(),
		postgresqlreceiver.NewFactory(),
		mongodbreceiver.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Exporters, err = exporter.MakeFactoryMap([]exporter.Factory{
		loggingexporter.NewFactory(),
		otlpexporter.NewFactory(),
		otlphttpexporter.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	factories.Processors, err = processor.MakeFactoryMap([]processor.Factory{
		// frontend.NewProcessorFactory(),
		batchprocessor.NewFactory(),
		filterprocessor.NewFactory(),
		memorylimiterprocessor.NewFactory(),
		resourceprocessor.NewFactory(),
		resourcedetectionprocessor.NewFactory(),
		attributesprocessor.NewFactory(),
	}...)
	if err != nil {
		return otelcol.Factories{}, err
	}

	return factories, nil
}

func (c *HostAgent) HasValidTags() bool {
	if c.hostTags == "" {
		return true
	}
	pairs := strings.Split(c.hostTags, ",")
	for _, pair := range pairs {
		keyValue := strings.Split(pair, ":")
		if len(keyValue) != 2 {
			return false
		}
	}
	return true
}
