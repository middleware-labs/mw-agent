package agent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testIMDSToken = "test-imds-token"

// newFakeIMDS serves IMDSv2 token and metadata requests, and points
// ec2MetadataBaseURL at itself for the duration of the test.
func newFakeIMDS(t *testing.T, metadataByPath map[string]string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /latest/api/token", func(responseWriter http.ResponseWriter, request *http.Request) {
		_, _ = responseWriter.Write([]byte(testIMDSToken))
	})
	mux.HandleFunc("GET /latest/meta-data/{path}", func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-aws-ec2-metadata-token") != testIMDSToken {
			responseWriter.WriteHeader(http.StatusUnauthorized)
			return
		}
		metadataValue, found := metadataByPath[request.PathValue("path")]
		if !found {
			responseWriter.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = responseWriter.Write([]byte(metadataValue))
	})
	fakeIMDSServer := httptest.NewServer(mux)
	t.Cleanup(fakeIMDSServer.Close)
	setEC2MetadataBaseURL(t, fakeIMDSServer.URL+"/latest")
}

// setEC2MetadataBaseURL points IMDS at baseURL and clears the cached EC2
// hostname, so each test starts as a freshly started agent.
func setEC2MetadataBaseURL(t *testing.T, baseURL string) {
	t.Helper()
	originalBaseURL := ec2MetadataBaseURL
	ec2MetadataBaseURL = baseURL
	clearCachedEC2Hostname()
	t.Cleanup(func() {
		ec2MetadataBaseURL = originalBaseURL
		clearCachedEC2Hostname()
	})
}

func clearCachedEC2Hostname() {
	cachedEC2Hostname.mu.Lock()
	defer cachedEC2Hostname.mu.Unlock()
	cachedEC2Hostname.hostname = ""
}

func unreachableIMDSURL(t *testing.T) string {
	t.Helper()
	unreachableServer := httptest.NewServer(http.NotFoundHandler())
	unreachableURL := unreachableServer.URL
	unreachableServer.Close()
	return unreachableURL + "/latest"
}

func TestGetHostnameForPlatformEC2KeepsCachedHostnameWhenIMDSFailsLater(t *testing.T) {
	t.Setenv(HostIDEnvVar, "")
	newFakeIMDS(t, map[string]string{"hostname": "ip-172-31-23-233.eu-north-1.compute.internal"})
	require.Equal(t, "ip-172-31-23-233.eu-north-1.compute.internal", GetHostnameForPlatform(InfraPlatformEC2))

	// Later calls must not fall back to os.Hostname() when IMDS stops answering.
	ec2MetadataBaseURL = unreachableIMDSURL(t)
	assert.Equal(t, "ip-172-31-23-233.eu-north-1.compute.internal", GetHostnameForPlatform(InfraPlatformEC2))
}

func TestGetHostnameForPlatformEC2RetriesIMDSAfterFailedRead(t *testing.T) {
	t.Setenv(HostIDEnvVar, "")
	setEC2MetadataBaseURL(t, unreachableIMDSURL(t))
	require.Equal(t, osHostname(t), GetHostnameForPlatform(InfraPlatformEC2))

	// The failed read is not cached: once IMDS answers, the EC2 name is used.
	newFakeIMDS(t, map[string]string{"hostname": "ip-172-31-23-233.eu-north-1.compute.internal"})
	assert.Equal(t, "ip-172-31-23-233.eu-north-1.compute.internal", GetHostnameForPlatform(InfraPlatformEC2))
}

func osHostname(t *testing.T) string {
	t.Helper()
	hostname, err := os.Hostname()
	require.NoError(t, err)
	return hostname
}

func TestGetHostnameForPlatformEC2ReturnsIMDSHostname(t *testing.T) {
	t.Setenv(HostIDEnvVar, "")
	newFakeIMDS(t, map[string]string{
		"hostname":       "ip-172-31-23-233.eu-north-1.compute.internal",
		"local-hostname": "should-not-be-used",
	})

	assert.Equal(t, "ip-172-31-23-233.eu-north-1.compute.internal", GetHostnameForPlatform(InfraPlatformEC2))
}

func TestGetHostnameForPlatformEC2FallsBackToOSHostname(t *testing.T) {
	t.Setenv(HostIDEnvVar, "")

	t.Run("metadata request fails", func(t *testing.T) {
		newFakeIMDS(t, map[string]string{})
		assert.Equal(t, osHostname(t), GetHostnameForPlatform(InfraPlatformEC2))
	})

	t.Run("metadata hostname is empty", func(t *testing.T) {
		newFakeIMDS(t, map[string]string{"hostname": "  \n"})
		assert.Equal(t, osHostname(t), GetHostnameForPlatform(InfraPlatformEC2))
	})

	t.Run("IMDS unreachable", func(t *testing.T) {
		unreachableServer := httptest.NewServer(http.NotFoundHandler())
		unreachableURL := unreachableServer.URL
		unreachableServer.Close()
		setEC2MetadataBaseURL(t, unreachableURL+"/latest")
		assert.Equal(t, osHostname(t), GetHostnameForPlatform(InfraPlatformEC2))
	})
}

func TestGetHostnameForPlatformNonEC2SkipsIMDS(t *testing.T) {
	t.Setenv(HostIDEnvVar, "")
	newFakeIMDS(t, map[string]string{"hostname": "ip-10-0-0-1.ec2.internal"})

	assert.Equal(t, osHostname(t), GetHostnameForPlatform(InfraPlatformInstance))
}

func TestGetHostnameForPlatformHostIDEnvWins(t *testing.T) {
	newFakeIMDS(t, map[string]string{"hostname": "ip-10-0-0-1.ec2.internal"})

	t.Setenv(HostIDEnvVar, " supervisor-resolved-host-id ")
	assert.Equal(t, "supervisor-resolved-host-id", GetHostnameForPlatform(InfraPlatformEC2))
	assert.Equal(t, "supervisor-resolved-host-id", GetHostnameForPlatform(InfraPlatformInstance))

	t.Setenv(HostIDEnvVar, "   ")
	assert.Equal(t, "ip-10-0-0-1.ec2.internal", GetHostnameForPlatform(InfraPlatformEC2))
}

func TestParseInfraPlatform(t *testing.T) {
	tests := []struct {
		platformName     string
		expectedPlatform InfraPlatform
		expectedOK       bool
	}{
		{"instance", InfraPlatformInstance, true},
		{"kubernetes", InfraPlatformKubernetes, true},
		{"ecsec2", InfraPlatformECSEC2, true},
		{"ecsfargate", InfraPlatformECSFargate, true},
		{"cycleio", InfraPlatformCycleIO, true},
		{"ec2", InfraPlatformEC2, true},
		{" EC2 ", InfraPlatformEC2, true},
		{"", InfraPlatformInstance, false},
		{"unknown", InfraPlatformInstance, false},
		{"gcp", InfraPlatformInstance, false},
	}
	for _, testCase := range tests {
		t.Run(testCase.platformName, func(t *testing.T) {
			infraPlatform, ok := ParseInfraPlatform(testCase.platformName)
			assert.Equal(t, testCase.expectedOK, ok)
			assert.Equal(t, testCase.expectedPlatform, infraPlatform)
		})
	}
}
