package healthcheck

import (
	"context"
	"fmt"
	"net"
	"strings"

	_ "github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

type HealthChecker interface {
	CheckHealth(ctx context.Context) error
}

func NewHealthChecker(receiverName string, rawCfg any) (HealthChecker, error) {
	switch {
	case strings.HasPrefix(receiverName, "postgresql"):
		var cfg PostgresqlReceiver
		if err := mapstructure.Decode(rawCfg, &cfg); err != nil {
			return nil, fmt.Errorf("postgresql: failed to decode config: %w", err)
		}
		if err := cfg.Validate(); err != nil {
			return nil, err
		}
		return &cfg, nil
	default:
		return nil, fmt.Errorf("healthcheck not implemented for receiver: %s", receiverName)
	}
}

func parseEndpoint(endpoint, defaultPort string) (host, port string) {
	h, p, err := net.SplitHostPort(endpoint)
	if err != nil {
		return endpoint, defaultPort
	}
	return h, p
}
