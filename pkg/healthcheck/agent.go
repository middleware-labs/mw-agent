package healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HealthCheckResponse struct {
	statusCode int
	StatusMsg  string    `json:"status"`
	UpSince    time.Time `json:"upSince"`
	Uptime     string    `json:"uptime"`
}

type AgentHealthChecker struct {
	endpoint string
}

func NewAgentHealthChecker(endpoint string) *AgentHealthChecker {
	return &AgentHealthChecker{endpoint: endpoint}
}

func (a *AgentHealthChecker) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("collector not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("collector unhealthy: status %d", resp.StatusCode)
	}

	return nil
}

func (a *AgentHealthChecker) GetStatus(ctx context.Context) (*HealthCheckResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("collector not reachable: %w", err)
	}
	defer resp.Body.Close()

	var hcResp HealthCheckResponse
	hcResp.statusCode = resp.StatusCode
	if err := json.NewDecoder(resp.Body).Decode(&hcResp); err != nil {
		return nil, fmt.Errorf("failed to parse health check response: %w", err)
	}

	return &hcResp, nil
}
