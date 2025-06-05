package utils

var CategorizedIntegrations = map[string][]string{
	"🗄️  Database": {
		"Postgres", "MongoDB", "MySQL", "MariaDB", "Redis",
		"Clickhouse", "Cassandra", "ElasticSearch", "OracleDB",
		"SQLServer", "MongoDBAtlas",
	},
	"📦 Logs": {
		"Journald", "WindowsEventLogs",
	},
	"📊 Telemetry": {
		"Prometheus",
	},
	"📡 Streaming": {
		"Redpanda", "Kafka", "RabbitMQ",
	},
	"🌐 Networking": {
		"Apache", "Nginx",
	},
	"🪟 Windows": {
		"WindowsEventLogs",
	},
}

type AgentConfig struct {
	APIKey string `yaml:"api-key"`
	Target string `yaml:"target"`
}
