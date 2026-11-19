package main

import (
	"testing"

	"github.com/middleware-labs/mw-agent/pkg/agent"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestDetectInfraPlatformUsesEnvOverride(t *testing.T) {
	observedCore, observedLogs := observer.New(zap.WarnLevel)
	logger := zap.New(observedCore)

	// AWS_EXECUTION_ENV would detect ECS EC2; the override must win.
	t.Setenv("AWS_EXECUTION_ENV", "AWS_ECS_EC2")
	t.Setenv(agent.InfraPlatformEnvVar, "ec2")

	assert.Equal(t, agent.InfraPlatformEC2, detectInfraPlatform(logger))
	assert.Equal(t, 0, observedLogs.Len())
}

func TestDetectInfraPlatformUnknownEnvWarnsAndDetects(t *testing.T) {
	observedCore, observedLogs := observer.New(zap.WarnLevel)
	logger := zap.New(observedCore)

	// AWS_EXECUTION_ENV makes detection return without probing IMDS.
	t.Setenv("AWS_EXECUTION_ENV", "AWS_ECS_FARGATE")
	t.Setenv(agent.InfraPlatformEnvVar, "not-a-platform")

	assert.Equal(t, agent.InfraPlatformECSFargate, detectInfraPlatform(logger))
	assert.Equal(t, 1, observedLogs.FilterMessage("ignoring unknown infra platform, detecting it instead").Len())
}
