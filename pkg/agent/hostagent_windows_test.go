//go:build windows

package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
)

func TestWindowsFactoriesIncludeSupervisorComponents(t *testing.T) {
	factories, err := (&HostAgent{}).getFactories()
	require.NoError(t, err)

	assert.Contains(t, factories.Extensions, component.MustNewType("opamp"))
	assert.Contains(t, factories.Receivers, component.MustNewType("nop"))
	assert.Contains(t, factories.Exporters, component.MustNewType("nop"))
}
