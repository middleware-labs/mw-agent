package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"gopkg.in/yaml.v2"
)

type ConfigType struct {
	Docker   map[string]interface{} `json:"docker"`
	NoDocker map[string]interface{} `json:"nodocker"`
}

type pgdbConfiguration struct {
	Path string `json:"path"`
}

type APIResponseForYAML struct {
	Status     bool              `json:"status"`
	Config     ConfigType        `json:"config"`
	PgdbConfig pgdbConfiguration `json:"pgdb_config"`
	Message    string            `json:"message"`
}

type APIResponseForRestart struct {
	Status  bool   `json:"status"`
	Restart bool   `json:"restart"`
	Message string `json:"message"`
}

var (
	apiURLForYAML    = "https://app.middleware.io/api/v1/agent/ingestion-rules/%s?config=%s&platform=linux&host_id=%s"
	apiURLForRestart = "https://app.middleware.io/api/v1/agent/restart-status/%s?platform=linux&host_id=%s"
)

const (
	dockerSocketPath  = "/var/run/docker.sock"
	yamlFile          = "configyamls/all/otel-config.yaml"
	yamlFileNoDocker  = "configyamls/nodocker/otel-config.yaml"
	recursiveInterval = 20 * time.Second
)

func checkForConfigURLOverrides() (string, string) {

	// fmt.Println("update-yaml.go: checkForConfigURLOverrides: MW_API_URL_FOR_RESTART", os.Getenv("MW_API_URL_FOR_RESTART"))
	// fmt.Println("update-yaml.go: checkForConfigURLOverrides: MW_API_URL_FOR_YAML", os.Getenv("MW_API_URL_FOR_YAML"))

	if os.Getenv("MW_API_URL_FOR_RESTART") != "" {
		apiURLForRestart = os.Getenv("MW_API_URL_FOR_RESTART")
	}

	if os.Getenv("MW_API_URL_FOR_YAML") != "" {
		apiURLForYAML = os.Getenv("MW_API_URL_FOR_YAML")
	}

	return apiURLForRestart, apiURLForYAML
}

func updatepgdbConfig(config map[string]interface{}, pgdbConfig pgdbConfiguration) map[string]interface{} {

	// Read the YAML file
	yamlData, err := ioutil.ReadFile(pgdbConfig.Path)
	if err != nil {
		log.Println("Failed to read YAML file: ", err)
	}

	// Unmarshal the YAML data into a temporary map[string]interface{}
	tempMap := make(map[string]interface{})
	err = yaml.Unmarshal(yamlData, &tempMap)
	if err != nil {
		log.Println("Failed to unmarshal YAML:", err)
	}

	// Add the temporary map to the existing "receiver" key
	receiverData, ok := config["receivers"].(map[string]interface{})
	if !ok {
		log.Println("Failed to access 'receivers' key in existing config")
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

	return config
}

func updateYAML(configType, yamlPath string) error {
	_, apiURLForYAML := checkForConfigURLOverrides()
	apiKey, ok := os.LookupEnv("MW_API_KEY")
	if !ok || apiKey == "" {
		return fmt.Errorf("MW_API_KEY not found in environment variables")
	}

	hostname := getHostname()

	// Call Webhook
	apiURL := fmt.Sprintf(apiURLForYAML, apiKey, configType, hostname)

	resp, err := http.Get(apiURL)
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
	var apiResponse APIResponseForYAML
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
	apiYAMLConfig = updatepgdbConfig(apiYAMLConfig, pgdbConfig)

	apiYAMLBytes, err := yaml.Marshal(apiYAMLConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal API YAML data: %v", err)
	}

	if err := os.WriteFile(yamlPath, apiYAMLBytes, 0644); err != nil {
		return fmt.Errorf("failed to write new YAML data to file: %v", err)
	}
	return nil
}

func getUpdatedYAMLPath() string {
	configType := "docker"
	yamlPath := yamlFile
	if !isSocket(dockerSocketPath) {
		configType = "nodocker"
		yamlPath = yamlFileNoDocker
		fmt.Println("Docker socket not found at", dockerSocketPath, ", using nodocker config")
	}

	if err := updateYAML(configType, yamlPath); err != nil {
		fmt.Println(fmt.Errorf("UpdateYAML error: %v", err))
	}

	return yamlPath
}

func restartHostAgent() error {
	getUpdatedYAMLPath()
	cmd := exec.Command("kill", "-SIGHUP", fmt.Sprintf("%d", os.Getpid()))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func callRestartStatusAPI() {

	// fmt.Println("Starting recursive restart check......")

	apiURLForRestart, _ := checkForConfigURLOverrides()

	apiKey, ok := os.LookupEnv("MW_API_KEY")
	if !ok || apiKey == "" {
		log.Println("MW_API_KEY not found in environment variables for recursive restart")
	}

	hostname := getHostname()

	// Prepare API URL
	apiURL := fmt.Sprintf(apiURLForRestart, apiKey, hostname)

	resp, err := http.Get(apiURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Printf("Failed to call Restart-API: Status-code %d with %v", resp.StatusCode, err)
	}
	defer resp.Body.Close()

	var apiResponse APIResponseForRestart
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Printf("Failed to unmarshal Restart-API response: %v", err)
	}

	// fmt.Println("Restart API Response: apiResponse.Restart", apiResponse.Restart)

	if apiResponse.Restart {
		log.Println("Restarting Linux Agent......")
		if err := restartHostAgent(); err != nil {
			log.Printf("Error restarting agent: %v", err)
		}
	}
}

func listenForConfigChanges() {

	// fmt.Println("listening for config changes")
	go func() {

		restartIntervalString, ok := os.LookupEnv("MW_CHECK_FOR_RESTART_INTERVAL")
		if restartIntervalString == "" || !ok {
			log.Println("MW_CHECK_FOR_RESTART_INTERVAL not found in environment variables for recursive restart")
			restartIntervalString = recursiveInterval.String()
		}

		restartInterval, okk := time.ParseDuration(restartIntervalString)
		if okk != nil {
			log.Println("Error parsing restart interval duration", restartIntervalString)
			restartInterval = recursiveInterval
		}

		log.Println("Check for Config Changes After ... ===>", restartInterval)

		for range time.Tick(restartInterval) {
			callRestartStatusAPI()
		}
	}()
}
