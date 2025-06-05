package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/middleware-labs/mw-agent/integrations/utils"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

func SaveAndSendConfig(integration string, config interface{}, hostname string, agentConfig utils.AgentConfig) {
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		fmt.Printf("Failed to marshal YAML: %v\n", err)
		return
	}

	baseDir := "/etc/mw-agent/integrations"
	err = os.MkdirAll(baseDir, 0755)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Println("❌ Permission denied. Please run this command with sudo:")
			fmt.Println("   sudo mw-agent integration")
		} else {
			fmt.Printf("Failed to create directories: %v\n", err)
		}
		return
	}

	filePath := filepath.Join(baseDir, integration+".yaml")
	err = os.WriteFile(filePath, yamlData, 0644)
	if err != nil {
		fmt.Printf("Failed to write config file: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("🎉 %s configuration completed!\n", cases.Title(language.English).String(integration))
	fmt.Printf("📁 File saved at: %s", filePath)

	u, err := url.Parse(agentConfig.Target)
	if err != nil {
		log.Fatal(err)
	}
	if u.Scheme == "https" && u.Port() == "443" {
		u.Host = u.Hostname() // remove the port
	}

	baseURL := fmt.Sprintf("%s/api/v1", u.String())
	accountToken := agentConfig.APIKey

	authData, err := FetchAuthData(baseURL, hostname, accountToken)
	if err != nil {
		fmt.Printf("❌ Auth fetch error: %v\n", err)
		return
	}

	sendIntegrationConfigToAPI(baseURL, filePath, hostname, authData, integration+"_config")

	fmt.Println("──────────────────────────────────────────────")
	color.Green("✅ Setup complete! You may now monitor %s.", cases.Title(language.English).String(integration))
	fmt.Println("──────────────────────────────────────────────")

}

func sendIntegrationConfigToAPI(baseURL, filePath, hostID string, authData *CaptureAuthData, configKey string) {
	const (
		timeZone  = "Asia/Kolkata"
		offset    = "+0530"
		sessionID = "4Vn39yx"
	)

	configPayload := map[string]interface{}{
		"linux": map[string]interface{}{
			"agent_restart_status": true,
			configKey: map[string]string{
				"path": filePath,
			},
		},
	}

	rawConfigJSON, err := json.Marshal(configPayload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal config: %v\n", err)
		return
	}

	encodedConfig := base64.StdEncoding.EncodeToString(rawConfigJSON)

	finalPayload := map[string]string{
		"value": encodedConfig,
	}

	finalJSON, err := json.Marshal(finalPayload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal final payload: %v\n", err)
		return
	}

	url := fmt.Sprintf("%s/agent/setting/by-identity/%s/%d/%d", baseURL, hostID, authData.AccountId, authData.ProjectId)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(finalJSON))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CLIENT_TIME_ZONE", timeZone)
	req.Header.Set("CLIENT_TIME_ZONE_OFFSET", offset)
	req.Header.Set("MW_USER_SESSION_ID", sessionID)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ API request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		return
	}

	var parsedResp map[string]interface{}
	if err := json.Unmarshal(body, &parsedResp); err != nil {
		fmt.Printf("❌ Failed to parse response JSON: %v\n", err)
		return
	}

	fmt.Println("📡 Configuration sent to middleware API")
	fmt.Println()
}
