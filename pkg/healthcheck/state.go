package healthcheck

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/goccy/go-yaml"
)

type AgentState struct {
	PID            int    `json:"pid"`
	OtelConfigFile string `json:"otel_config_file"`
}

func statePath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "mw-agent", "agent.state")
	default:
		return "/var/run/mw-agent/agent.state"
	}
}

func WriteState(otelConfigFile string) error {
	path := statePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	state := AgentState{
		PID:            os.Getpid(),
		OtelConfigFile: otelConfigFile,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return os.Rename(tmp, path)
}

func ReadState() (*AgentState, error) {
	data, err := os.ReadFile(statePath())
	if err != nil {
		return nil, fmt.Errorf("agent state not found (is the agent running?): %w", err)
	}

	var state AgentState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse agent state: %w", err)
	}

	return &state, nil
}

func RemoveState() error {
	return os.Remove(statePath())
}

func LoadReceivers(otelConfigFile string) (map[string]interface{}, error) {
	data, err := os.ReadFile(otelConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read otel config %s: %w", otelConfigFile, err)
	}

	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse otel config: %w", err)
	}

	receivers, ok := cfg["receivers"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no receivers found in config")
	}

	return receivers, nil
}
