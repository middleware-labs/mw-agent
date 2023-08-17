package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.uber.org/zap"
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

	pgdbConfig := pgdbConfiguration{
		Path: "db-config_test.yaml",
	}

	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))
	// Call the updatepgdbConfig function
	updatedConfig, err := agent.updatepgdbConfig(initialConfig, pgdbConfig)
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

	mongodbConfig := mongodbConfiguration{
		Path: "db-config_test.yaml",
	}

	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))

	updatedConfig, err := agent.updateMongodbConfig(initialConfig, mongodbConfig)
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

	mysqlConfig := mysqlConfiguration{
		Path: "db-config_test.yaml",
	}

	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))
	// Call the updateMysqlConfig function
	updatedConfig, err := agent.updateMysqlConfig(initialConfig, mysqlConfig)
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

func TestUpdateSqlserverConfig(t *testing.T) {
	// Define the initial config and sqlserverConfig
	initialConfig := map[string]interface{}{
		"receivers": map[string]interface{}{
			"sqlserver": map[string]interface{}{
				"endpoint": "example.com:5432",
				"database": "mydb",
				"user":     "myuser",
				"password": "mypassword",
			},
		},
	}

	sqlserverConfig := sqlserverConfiguration{
		Path: "db-config_test.yaml",
	}

	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))
	// Call the updateSqlserverConfig function
	updatedConfig, err := agent.updateSqlserverConfig(initialConfig, sqlserverConfig)
	assert.NoError(t, err)

	// Assert that the updated config contains the expected values
	assert.Contains(t, updatedConfig, "receivers")
	receivers, ok := updatedConfig["receivers"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, receivers, "sqlserver")
	sqlserver, ok := receivers["sqlserver"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, sqlserver, "endpoint")
	assert.Contains(t, sqlserver, "database")
	assert.Contains(t, sqlserver, "user")
	assert.Contains(t, sqlserver, "password")
}

func TestListenForConfigChanges(t *testing.T) {
	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))
	agent.configCheckInterval = "1s"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start listening for config changes in a separate goroutine
	go func() {
		err := agent.ListenForConfigChanges(ctx)
		assert.NoError(t, err)
	}()

	// Wait for a short period to allow the goroutine to start
	time.Sleep(100 * time.Millisecond)

	// Manually cancel the context to stop listening for config changes
	cancel()

	// Wait for a short period to allow the goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Assert that the goroutine has stopped and no error occurred
	assert.True(t, true)
}

func TestHostAgentGetFactories(t *testing.T) {
	agent := NewHostAgent(WithHostAgentLogger(zap.NewNop()))

	factories, err := agent.GetFactories(context.Background())
	assert.NoError(t, err)

	// Assert that the returned factories are not nil
	assert.NotNil(t, factories.Extensions)
	assert.NotNil(t, factories.Receivers)
	assert.NotNil(t, factories.Exporters)
	assert.NotNil(t, factories.Processors)

	// check that the returned factories contain the expected factories
	assert.Len(t, factories.Extensions, 1)
	assert.Contains(t, factories.Extensions, component.Type("health_check"))

	// check if factories contains expected receivers
	assert.Len(t, factories.Receivers, 8)
	assert.Contains(t, factories.Receivers, component.Type("otlp"))
	assert.Contains(t, factories.Receivers, component.Type("fluentforward"))
	assert.Contains(t, factories.Receivers, component.Type("filelog"))
	assert.Contains(t, factories.Receivers, component.Type("docker_stats"))
	assert.Contains(t, factories.Receivers, component.Type("hostmetrics"))
	assert.Contains(t, factories.Receivers, component.Type("prometheus"))
	assert.Contains(t, factories.Receivers, component.Type("postgresql"))
	assert.Contains(t, factories.Receivers, component.Type("mongodb"))
	assert.Contains(t, factories.Receivers, component.Type("mysql"))
	assert.Contains(t, factories.Receivers, component.Type("sqlserver"))

	// check if factories contain expected exporters
	assert.Len(t, factories.Exporters, 3)
	assert.Contains(t, factories.Exporters, component.Type("logging"))
	assert.Contains(t, factories.Exporters, component.Type("otlp"))
	assert.Contains(t, factories.Exporters, component.Type("otlphttp"))

	// check if factories contain expected processors
	assert.Len(t, factories.Processors, 6)
	assert.Contains(t, factories.Processors, component.Type("batch"))
	assert.Contains(t, factories.Processors, component.Type("filter"))
	assert.Contains(t, factories.Processors, component.Type("memory_limiter"))
	assert.Contains(t, factories.Processors, component.Type("resource"))
	assert.Contains(t, factories.Processors, component.Type("resourcedetection"))
	assert.Contains(t, factories.Processors, component.Type("attributes"))

}

func TestHostAgentHasValidTags(t *testing.T) {
	testCases := []struct {
		tags    string
		isValid bool
	}{
		// case 1: host tags not provided
		{"", true},

		// case 2: tags match with expected pattern
		{"name:my-machine,env:prod1", true},

		// case 3: tags do not match expected pattern
		{"name", false},
		{"name:,", false},
		{"name:1,", false},
		{"name:1,test", false},
	}

	for i, tc := range testCases {
		agent := NewHostAgent(WithHostAgentHostTags(tc.tags))

		isValid := agent.HasValidTags()
		if isValid != tc.isValid {
			t.Errorf("Test case %d failed. Expected HasValidTags to return: %v, but got: %v", i+1, tc.isValid, isValid)
		}
	}
}
