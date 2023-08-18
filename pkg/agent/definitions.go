package agent

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"go.opentelemetry.io/collector/otelcol"
)

// Agent interface provides common methods for different agents
// like host agent (Linux, Windows) & Kubernetes
type Agent interface {
	GetFactories(ctx context.Context) (otelcol.Factories, error)
	GetUpdatedYAMLPath(ctx context.Context) (string, error)
	ListenForConfigChanges(ctx context.Context) error
}

// Config stores general configuration for all agent types
type Config struct {
	ApiKey                    string
	Target                    string
	ConfigCheckInterval       string
	ApiURLForConfigCheck      string
	HostTags                  string
	DockerEndpoint            string
	EnableSyntheticMonitoring bool
	Logfile                   string
	LogfileSize               int
}

// String() implements stringer interface for Config
func (c Config) String() string {
	var s string
	s += fmt.Sprintf("api-key: %s, ", c.ApiKey)
	s += fmt.Sprintf("target: %s, ", c.Target)
	s += fmt.Sprintf("config-check-interval: %s, ", c.ConfigCheckInterval)
	s += fmt.Sprintf("api-url-for-config-check: %s, ", c.ApiURLForConfigCheck)
	s += fmt.Sprintf("host-tags: %s, ", c.HostTags)
	s += fmt.Sprintf("docker-endpoint: %s, ", c.DockerEndpoint)
	s += fmt.Sprintf("enable-synthetic-monitoring: %t, ", c.EnableSyntheticMonitoring)
	s += fmt.Sprintf("logfile: %s, ", c.Logfile)
	s += fmt.Sprintf("logfile-size: %d", c.LogfileSize)
	return s
}

func isSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

var isSocketFn func(path string) bool = isSocket

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
