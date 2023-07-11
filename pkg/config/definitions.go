package config

import (
	"context"
	"io/fs"
	"os"

	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

func SetLogger(l *zap.Logger) {
	logger = l
}

type Config interface {
	GetFactories(ctx context.Context) (otelcol.Factories, error)
	GetUpdatedYAMLPath(ctx context.Context) (string, error)
	ListenForConfigChanges(ctx context.Context) error
	SetLogger(ctx context.Context, l *zap.Logger)
}

func isSocket(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.Mode().Type() == fs.ModeSocket
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}
