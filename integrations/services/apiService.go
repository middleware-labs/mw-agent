package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func SendIntegrationConfigToAPI(baseURL, filePath, hostID string, authData *CaptureAuthData, configKey string) {
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

	url := fmt.Sprintf("%s/agent/setting/withoutAuth/%s/%d/%d", baseURL, hostID, authData.AccountId, authData.ProjectId)

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
