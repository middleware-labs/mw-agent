package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponents(t *testing.T) {
	factories, err := Components()
	require.NoError(t, err)

	exporters := factories.Exporters
	receivers := factories.Receivers
	processors := factories.Processors

	assert.True(t, exporters["otlp"] != nil)

	assert.True(t, receivers["otlp"] != nil)

	assert.True(t, processors["memory_limiter"] != nil)
}
