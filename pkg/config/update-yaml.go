package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gopkg.in/yaml.v2"
)

var (
	ErrRestartStatusAPINotOK = errors.New("received error code from the server")
	ErrReceiverKeyNotFound   = errors.New("'receivers' key not found")
)

type configType struct {
	Docker   map[string]interface{} `json:"docker"`
	NoDocker map[string]interface{} `json:"nodocker"`
}

type pgdbConfiguration struct {
	Path string `json:"path"`
}

type apiResponseForYAML struct {
	Status     bool              `json:"status"`
	Config     configType        `json:"config"`
	PgdbConfig pgdbConfiguration `json:"pgdb_config"`
	Message    string            `json:"message"`
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
	dockerSocketPath = "/var/run/docker.sock"
	yamlFile         = "configyamls/all/otel-config.yaml"
	yamlFileNoDocker = "configyamls/nodocker/otel-config.yaml"
)

/*func (c *Config) checkForConfigURLOverrides() (string, string) {

	if os.Getenv("MW_API_URL_FOR_RESTART") != "" {
		apiURLForRestart = os.Getenv("MW_API_URL_FOR_RESTART")
	}

	if os.Getenv("MW_API_URL_FOR_YAML") != "" {
		apiURLForYAML = os.Getenv("MW_API_URL_FOR_YAML")
	}

	return apiURLForRestart, apiURLForYAML
}*/

func updatepgdbConfig(config map[string]interface{}, pgdbConfig pgdbConfiguration) (map[string]interface{}, error) {

	// Read the YAML file
	yamlData, err := ioutil.ReadFile(pgdbConfig.Path)
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
							fmt.Println("Invalid key type:", k)
						}
					}
					receiverData[key] = oldMapValue
				}
			}
		}
	}

	return config, nil
}

func (c *Config) updateYAML(configType, yamlPath string) error {
	// _, apiURLForYAML := checkForConfigURLOverrides()

	hostname := getHostname()

	// Call Webhook
	u, err := url.Parse(c.ApiURLForConfigCheck)
	if err != nil {
		return err
	}

	baseUrl := u.JoinPath(apiPathForYAML).JoinPath(c.MWApiKey)
	params := url.Values{}
	params.Add("config", configType)
	params.Add("platform", runtime.GOOS)
	params.Add("host_id", hostname)
	// Add Query Parameters to the URL
	baseUrl.RawQuery = params.Encode() // Escape Query Parameters

	//apiURL := fmt.Sprintf(apiURLForYAML, c.MWApiKey, configType, hostname)
	resp, err := http.Get(baseUrl.String())
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to call YAML API: Status-code %d with %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read resp.Body response: %v", err)
	}

	// Unmarshal JSON response into ApiResponse struct
	var apiResponse apiResponseForYAML
	// fmt.Println("body: ", string(body))
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("failed to unmarshal API response: %v", err)
	}

	// Verify API Response
	if !apiResponse.Status {
		return fmt.Errorf("API response indicates failure: %v", apiResponse)
	}

	var apiYAMLConfig map[string]interface{}
	if len(apiResponse.Config.Docker) == 0 && len(apiResponse.Config.NoDocker) == 0 {
		return fmt.Errorf("API YAML JSON is either empty or not found: %v", apiResponse)
	} else {
		if configType == "docker" {
			apiYAMLConfig = apiResponse.Config.Docker
		} else {
			apiYAMLConfig = apiResponse.Config.NoDocker
		}
	}

	pgdbConfig := apiResponse.PgdbConfig
	apiYAMLConfig, err = updatepgdbConfig(apiYAMLConfig, pgdbConfig)
	if err != nil {
		return err
	}

	apiYAMLBytes, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal API YAML data: %v", err)
	}

	if err := os.WriteFile(yamlPath, apiYAMLBytes, 0644); err != nil {
		return fmt.Errorf("failed to write new YAML data to file: %v", err)
	}
	return nil
}

func (c *Config) GetUpdatedYAMLPath() (string, error) {
	configType := "docker"
	yamlPath := yamlFile
	if !isSocket(dockerSocketPath) {
		configType = "nodocker"
		yamlPath = yamlFileNoDocker
		fmt.Println("Docker socket not found at", dockerSocketPath, ", using nodocker config")
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

func (c *Config) callRestartStatusAPI() error {

	// fmt.Println("Starting recursive restart check......")
	// apiURLForRestart, _ := checkForConfigURLOverrides()
	hostname := getHostname()
	u, err := url.Parse(c.ApiURLForConfigCheck)
	if err != nil {
		return err
	}

	baseUrl := u.JoinPath(apiPathForRestart)
	baseUrl = baseUrl.JoinPath(c.MWApiKey)
	params := url.Values{}
	params.Add("host_id", hostname)
	params.Add("platform", runtime.GOOS)

	// Add Query Parameters to the URL
	baseUrl.RawQuery = params.Encode() // Escape Query Parameters

	// Prepare API URL
	apiURL := fmt.Sprintf(apiPathForRestart, c.MWApiKey, hostname)

	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to call Restart-API: Status-code %d with %v", resp.StatusCode, err)
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to call Restart-API: Status-code %d", resp.StatusCode)
		return ErrRestartStatusAPINotOK
	}

	var apiResponse apiResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Printf("Failed to unmarshal Restart-API response: %v", err)
		return err
	}

	if apiResponse.Restart {
		log.Println("Restarting Linux Agent......")
		if err := restartHostAgent(); err != nil {
			log.Printf("Error restarting agent: %v", err)
		}
	}

	return err
}

func (c *Config) ListenForConfigChanges(ctx context.Context) error {

	restartInterval, err := time.ParseDuration(c.ConfigCheckInterval)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(restartInterval)

	go func() {
		for {
			log.Println("Check for Config Changes After ... ===>", restartInterval)
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

func isSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
