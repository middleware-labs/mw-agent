package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAPIURLForConfigCheck(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedResult string
		err            error
	}{
		{
			name:           "URL with both '/' and '.' and without trailing '/'",
			url:            "https://myaccount.middleware.io",
			expectedResult: "https://app.middleware.io",
			err:            nil,
		},
		{
			name:           "URL with trailing '/'",
			url:            "https://myaccount.middleware.io/",
			expectedResult: "https://app.middleware.io",
			err:            nil,
		},
		{
			name:           "URL with only one '.'",
			url:            "https://middleware.io",
			expectedResult: "",
			err:            ErrInvalidTarget,
		},
		{
			name:           "URL with custom domain",
			url:            "https://myaccount.test.mw.io",
			expectedResult: "https://app.test.mw.io",
			err:            nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetAPIURLForConfigCheck(test.url)
			assert.Equal(t, test.err, err)
			assert.Equal(t, test.expectedResult, result)
		})
	}
}
