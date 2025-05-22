package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Struct for Auth API response
type CaptureAuthData struct {
	Account    string      `json:"account"`
	AccountId  int         `json:"account_id"`
	ProjectId  int         `json:"project_id"`
	ProjectUid string      `json:"project_uid"`
	Db         string      `json:"db"`
	Storage    string      `json:"storage"`
	Expires    int         `json:"expires"`
	Email      string      `json:"email"`
	Status     string      `json:"status"`
	BillMap    interface{} `json:"bill_map"` // You can use a specific type if the structure of BillMap is known
}

// FetchAuthData fetches auth data from the capture auth API with a Bearer token.
func FetchAuthData(baseURL, hostname, accountToken string) (*CaptureAuthData, error) {
	authURL := fmt.Sprintf("%s/auth", baseURL)

	req, err := http.NewRequest("POST", authURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth request: %v", err)
	}

	// Set the Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accountToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response: %v", err)
	}

	// Temporary struct to extract only the "data" field
	var apiResp struct {
		Data CaptureAuthData `json:"data"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth data: %v", err)
	}

	return &apiResp.Data, nil
}
