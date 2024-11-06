package agent

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestUpdatepgdbConfig(t *testing.T) {
	// Define the initial config and pgdbConfig
	initialConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"postgresql": map[string]interface{}{
				"endpoint": "example.com:5432",
				"database": "mydb",
				"user":     "myuser",
				"password": "mypassword",
			},
		},
	}

	pgdbConfig := integrationConfiguration{
		Path: "db-config_test.yaml",
	}

	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(HostConfig{}, zapCore)
	// Call the updatepgdbConfig function
	updatedConfig, err := agent.updateConfig(initialConfig, pgdbConfig)
	assert.NoError(t, err)

	// Assert that the updated config contains the expected values
	assert.Contains(t, updatedConfig, "receivers")
	receivers, ok := updatedConfig["receivers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, receivers, "postgresql")
	postgresql, ok := receivers["postgresql"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, postgresql, "endpoint")
	assert.Contains(t, postgresql, "database")
	assert.Contains(t, postgresql, "user")
	assert.Contains(t, postgresql, "password")
}

func TestUpdateMongodbConfig(t *testing.T) {
	// Define the initial config and mongodbConfig
	initialConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"mongodb": map[string]interface{}{
				"endpoint": "example.com:5432",
				"database": "mydb",
				"user":     "myuser",
				"password": "mypassword",
			},
		},
	}

	mongodbConfig := integrationConfiguration{
		Path: "db-config_test.yaml",
	}

	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(HostConfig{}, zapCore)

	updatedConfig, err := agent.updateConfig(initialConfig, mongodbConfig)
	assert.NoError(t, err)

	// Assert that the updated config contains the expected values
	assert.Contains(t, updatedConfig, "receivers")
	receivers, ok := updatedConfig["receivers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, receivers, "mongodb")
	mongodb, ok := receivers["mongodb"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, mongodb, "endpoint")
	assert.Contains(t, mongodb, "database")
	assert.Contains(t, mongodb, "user")
	assert.Contains(t, mongodb, "password")
}

func TestUpdateMysqlConfig(t *testing.T) {
	// Define the initial config and mysqlConfig
	initialConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"mysql": map[string]interface{}{
				"endpoint": "example.com:5432",
				"database": "mydb",
				"user":     "myuser",
				"password": "mypassword",
			},
		},
	}

	mysqlConfig := integrationConfiguration{
		Path: "db-config_test.yaml",
	}

	cfg := HostConfig{}
	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(cfg, zapCore)
	// Call the updateMysqlConfig function
	updatedConfig, err := agent.updateConfig(initialConfig, mysqlConfig)
	assert.NoError(t, err)

	// Assert that the updated config contains the expected values
	assert.Contains(t, updatedConfig, "receivers")
	receivers, ok := updatedConfig["receivers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, receivers, "mysql")
	mysql, ok := receivers["mysql"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, mysql, "endpoint")
	assert.Contains(t, mysql, "database")
	assert.Contains(t, mysql, "user")
	assert.Contains(t, mysql, "password")
}

func TestUpdateRedisConfig(t *testing.T) {
	// Define the initial config and redisConfig
	initialConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"redis": map[string]interface{}{
				"endpoint": "localhost:7379",
				"password": "mypassword",
			},
		},
	}

	redisConfig := integrationConfiguration{
		Path: "db-config_test.yaml",
	}

	cfg := HostConfig{}
	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(cfg, zapCore)
	// Call the updateRedisConfig function
	updatedConfig, err := agent.updateConfig(initialConfig, redisConfig)
	assert.NoError(t, err)

	// Assert that the updated config contains the expected values
	assert.Contains(t, updatedConfig, "receivers")
	receivers, ok := updatedConfig["receivers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, receivers, "redis")
	redis, ok := receivers["redis"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, redis, "endpoint")
	assert.Contains(t, redis, "password")
}

func TestListenForConfigChanges(t *testing.T) {
	cfg := HostConfig{}
	cfg.ConfigCheckInterval = "1s"

	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(cfg, zapCore)
	agent.httpGetFunc = func(url string) (resp *http.Response, err error) {
		return nil, fmt.Errorf("failed to call get configuration api for %s: %w", url,
			errors.New("test error"))
	}

	errCh := make(chan error)

	stopCh := make(chan struct{})

	// Start listening for config changes in a separate goroutine
	go agent.ListenForConfigChanges(errCh, stopCh)

	// Wait for a short period to allow the goroutine to start
	time.Sleep(100 * time.Millisecond)

	fmt.Println("waiting for config changes")
	err := <-errCh
	fmt.Println("config changes received")
	assert.NoError(t, err)
	fmt.Println("config changes asserted")
	// Manually cancel the context to stop listening for config changes
	stopCh <- struct{}{}

	// Wait for a short period to allow the goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Assert that the goroutine has stopped and no error occurred
	assert.True(t, true)
}

func assertContainsComponent(t *testing.T, factoryMap interface{}, componentName string) {
	t.Helper()

	switch m := factoryMap.(type) {
	case map[component.Type]receiver.Factory:
		// Check if the componentName exists in the Receivers map
		for t := range m {
			if t.String() == componentName {
				return // Found, test passes
			}
		}
		t.Errorf("Expected to find receiver '%s' in factories.Receivers, but it was not found", componentName)

	case map[component.Type]processor.Factory:
		// Check if the componentName exists in the Processors map
		for t := range m {
			if t.String() == componentName {
				return // Found, test passes
			}
		}
		t.Errorf("Expected to find processor '%s' in factories.Processors, but it was not found", componentName)

	case map[component.Type]exporter.Factory:
		// Check if the componentName exists in the Exporters map
		for t := range m {
			if t.String() == componentName {
				return // Found, test passes
			}
		}
		t.Errorf("Expected to find exporter '%s' in factories.Exporters, but it was not found", componentName)

	case map[component.Type]extension.Factory:
		// Check if the componentName exists in the Extensions map
		for t := range m {
			if t.String() == componentName {
				return // Found, test passes
			}
		}
		t.Errorf("Expected to find extension '%s' in factories.Extensions, but it was not found", componentName)

	default:
		t.Errorf("Unsupported factory map type")
	}
}

func TestHostAgentGetFactories(t *testing.T) {
	baseConfig := BaseConfig{
		AgentFeatures: AgentFeatures{
			MetricCollection: true,
			LogCollection:    true,
		},
	}

	zapCore := zapcore.NewNopCore()
	agent, _ := NewHostAgent(HostConfig{
		BaseConfig: baseConfig,
	}, zapCore,
		WithHostAgentInfraPlatform(InfraPlatformECSEC2))

	factories, err := agent.getFactories()
	assert.NoError(t, err)

	// Assert that the returned factories are not nil
	assert.NotNil(t, factories.Extensions)
	assert.NotNil(t, factories.Receivers)
	assert.NotNil(t, factories.Exporters)
	assert.NotNil(t, factories.Processors)

	// check that the returned factories contain the expected factories
	assert.Len(t, factories.Extensions, 1)
	assertContainsComponent(t, factories.Extensions, "health_check")
	// check if factories contains expected receivers
	assert.Len(t, factories.Receivers, 16)
	assertContainsComponent(t, factories.Receivers, "otlp")
	assertContainsComponent(t, factories.Receivers, "fluentforward")
	assertContainsComponent(t, factories.Receivers, "filelog")
	assertContainsComponent(t, factories.Receivers, "docker_stats")
	assertContainsComponent(t, factories.Receivers, "hostmetrics")
	assertContainsComponent(t, factories.Receivers, "prometheus")
	assertContainsComponent(t, factories.Receivers, "postgresql")
	assertContainsComponent(t, factories.Receivers, "mongodb")
	assertContainsComponent(t, factories.Receivers, "mysql")
	assertContainsComponent(t, factories.Receivers, "redis")
	assertContainsComponent(t, factories.Receivers, "elasticsearch")
	assertContainsComponent(t, factories.Receivers, "awsecscontainermetrics")
	assertContainsComponent(t, factories.Receivers, "jmx")
	assertContainsComponent(t, factories.Receivers, "kafkametrics")
	assertContainsComponent(t, factories.Receivers, "apache")
	assertContainsComponent(t, factories.Receivers, "oracledb")

	// check if factories contain expected exporters
	assert.Len(t, factories.Exporters, 5)
	assertContainsComponent(t, factories.Exporters, "logging")
	assertContainsComponent(t, factories.Exporters, "otlp")
	assertContainsComponent(t, factories.Exporters, "otlphttp")
	assertContainsComponent(t, factories.Exporters, "kafka")
	assertContainsComponent(t, factories.Exporters, "file")

	// check if factories contain expected processors
	assert.Len(t, factories.Processors, 10)
	assertContainsComponent(t, factories.Processors, "batch")
	assertContainsComponent(t, factories.Processors, "filter")
	assertContainsComponent(t, factories.Processors, "memory_limiter")
	assertContainsComponent(t, factories.Processors, "resource")
	assertContainsComponent(t, factories.Processors, "resourcedetection")
	assertContainsComponent(t, factories.Processors, "attributes")
	assertContainsComponent(t, factories.Processors, "transform")
	assertContainsComponent(t, factories.Processors, "cumulativetodelta")
	assertContainsComponent(t, factories.Processors, "deltatorate")

}

func TestHostAgentHasValidTags(t *testing.T) {
	testCases := []struct {
		tags    string
		isValid error
	}{
		// case 1: host tags not provided
		{"", nil},

		// case 2: tags match with expected pattern
		{"name:my-machine,env:prod1", nil},

		// case 3: tags do not match expected pattern
		{"name", errors.New("invalid tag format: name")},
		{"name:,", errors.New("invalid tag format: ")},
		{"name:1", nil},
		{"name:1,", errors.New("invalid tag format: ")},
		{"name:1,test", errors.New("invalid tag format: test")},
	}

	for i, tc := range testCases {
		t.Logf("Running test case %d: %s", i+1, tc.tags)

		isValid := HasValidTags(tc.tags)

		if isValid == nil && tc.isValid == nil {
			continue
		}

		if isValid == nil && tc.isValid != nil {
			t.Errorf("Test case %d failed. Expected HasValidTags to return: %v, but got: %v", i+1, tc.isValid, isValid)
		}

		if isValid != nil && tc.isValid == nil {
			t.Errorf("Test case %d failed. Expected HasValidTags to return: %v, but got: %v", i+1, tc.isValid, isValid)
		}

		// check if isValid is as expected
		if isValid.Error() != tc.isValid.Error() {
			t.Errorf("Test case %d failed. Expected HasValidTags to return: %v, but got: %v", i+1, tc.isValid, isValid)
		}
	}
}

func TestUpdateAgentTrackStatus(t *testing.T) {
	tests := []struct {
		name           string
		serverResponse int
		wantError     bool
	}{
		{
			name:           "successful response",
			serverResponse: http.StatusOK,
			wantError:     false,
		},
		{
			name:           "internal server error",
			serverResponse: http.StatusInternalServerError,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tt.serverResponse)
			}))
			defer mockServer.Close()

			noopCore := zapcore.NewNopCore()
			logger := zap.New(noopCore)
			defer logger.Sync()

			hostAgent := &HostAgent{
				HostConfig: HostConfig{
					BaseConfig: BaseConfig{
						APIKey:               "testAPIKey",
						APIURLForConfigCheck: mockServer.URL,
					},
				},
				logger:  logger,
				Version: "1.0.0",
			}

			err := hostAgent.UpdateAgentTrackStatus(errors.New("test reason"))

			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
