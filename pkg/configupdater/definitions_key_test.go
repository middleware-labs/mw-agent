package configupdater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigStringHidesAPIKey(t *testing.T) {
	const apiKey = "secret-api-key"
	base := BaseConfig{APIKey: apiKey}

	assert.NotContains(t, base.String(), apiKey)
	assert.NotContains(t, KubeAgent{BaseConfig: base}.String(), apiKey)
}
