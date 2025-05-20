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

func SendPostgresConfigToAPI(filePath string) {
	const (
		hostID    = "" // TODO: Add the host ID of your agent
		baseURL   = "" // TODO: Add the base URL of your API (Local or Deployed)
		timeZone  = "Asia/Kolkata"
		offset    = "+0530"
		sessionID = "4Vn39yx"
	)

	// Directly define the config payload as expected by backend
	configPayload := map[string]interface{}{
		"linux": map[string]interface{}{
			"agent_restart_status": true,
			"postgres_config": map[string]string{
				"path": filePath,
			},
		},
	}

	// Marshal full Payload.Config to JSON
	rawConfigJSON, err := json.Marshal(configPayload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal config: %v\n", err)
		return
	}

	// Base64 encode the config
	encodedConfig := base64.StdEncoding.EncodeToString(rawConfigJSON)

	// Wrap it inside { "value": "<base64_string>" }
	finalPayload := map[string]string{
		"value": encodedConfig,
	}

	// Marshal the final payload
	finalJSON, err := json.Marshal(finalPayload)
	if err != nil {
		fmt.Printf("❌ Failed to marshal final payload: %v\n", err)
		return
	}

	// Prepare request
	url := fmt.Sprintf("%s%s", baseURL, hostID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(finalJSON))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return
	}

	// Set required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CLIENT_TIME_ZONE", timeZone)
	req.Header.Set("CLIENT_TIME_ZONE_OFFSET", offset)
	req.Header.Set("MW_USER_SESSION_ID", sessionID)
	// req.Header.Set("Authorization", "Bearer <your_token>") // Add this if needed

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ API request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Step 5: Read & print response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		return
	}

	// fmt.Printf("📡 Status: %s\n", resp.Status)
	// fmt.Println("📨 Response Body:")
	// fmt.Println(string(body))

	// Step 6: Verify path was stored correctly
	var parsedResp struct {
		Setting struct {
			Config struct {
				Linux struct {
					PostgresConfig struct {
						Path string `json:"path"`
					} `json:"postgres_config"`
				} `json:"linux"`
			} `json:"config"`
		} `json:"setting"`
		Status bool `json:"status"`
	}

	if err := json.Unmarshal(body, &parsedResp); err != nil {
		fmt.Printf("❌ Failed to parse response JSON: %v\n", err)
		return
	}

	if parsedResp.Status && parsedResp.Setting.Config.Linux.PostgresConfig.Path == filePath {
		fmt.Println("✅ Postgres config path verified successfully in the DB.")
	} else {
		fmt.Println("⚠️ Config path was not stored correctly or response format changed.")
	}
}
