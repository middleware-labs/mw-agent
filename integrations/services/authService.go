package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Struct for Auth API response
type CaptureAuthData struct {
	Alias      string `json:"alias"`
	AccountUid string `json:"accountUid"`
	AccountId  int    `json:"accountId"`
	ProjectId  int    `json:"projectId"`
	ProjectUid string `json:"projectUid"`
	Db         string `json:"db"`
	Storage    string `json:"storage"`
	Expires    int    `json:"expires"`
	Email      string `json:"email"`
	Status     string `json:"status"`
}

// Fetch auth data from GitHub auth-config API
func FetchAuthData(baseURL, hostname, accountToken string) (*CaptureAuthData, error) {
	authURL := fmt.Sprintf("%s/integration/github/auth-config/%s/%s", baseURL, hostname, accountToken)

	resp, err := http.Get(authURL)
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth response: %v", err)
	}

	var authData CaptureAuthData
	if err := json.Unmarshal(body, &authData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth data: %v", err)
	}

	return &authData, nil
}
