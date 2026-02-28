package agent

// ========== Integration Configuration ==========
type IntegrationConfig struct {
	Postgresql *Postgresql `json:"postgresql,omitempty" yaml:"postgresql,omitempty"`
}

// ========= Integration Interface ==========
type Integration interface {
	UpdateReceiverConfig(receiverConfig map[string]interface{}) error
}

// ========== Postgresql Configuration ==========
type Postgresql struct {
	Username              string                           `json:"username,omitempty" yaml:"username,omitempty"`
	Password              string                           `json:"password,omitempty" yaml:"password,omitempty"`
	Endpoint              string                           `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Databases             []string                         `json:"databases,omitempty" yaml:"databases,omitempty"`
	ExcludeDatabases      []string                         `json:"exclude_databases,omitempty" yaml:"exclude_databases,omitempty"`
	CollectionInterval    string                           `json:"collection_interval,omitempty" yaml:"collection_interval,omitempty"`
	Transport             string                           `json:"transport,omitempty" yaml:"transport,omitempty"`
	TLS                   *PostgresqlTLS                   `json:"tls,omitempty" yaml:"tls,omitempty"`
	Events                *PostgresqlEvents                `json:"events,omitempty" yaml:"events,omitempty"`
	QuerySampleCollection *PostgresqlQuerySampleCollection `json:"query_sample_collection,omitempty" yaml:"query_sample_collection,omitempty"`
	TopQueryCollection    *PostgresqlTopQueryCollection    `json:"top_query_collection,omitempty" yaml:"top_query_collection,omitempty"`
	ConectionPool         *PostgresqlConnectionPool        `json:"connection_pool,omitempty" yaml:"connection_pool,omitempty"`
}
type PostgresqlTLS struct {
	Insecure           bool   `json:"insecure,omitempty" yaml:"insecure,omitempty"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"`
	CertFile           string `json:"cert_file,omitempty" yaml:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty" yaml:"key_file,omitempty"`
	CaFile             string `json:"ca_file,omitempty" yaml:"ca_file,omitempty"`
}
type PostgresqlEvents struct {
	DbServerQuerySample PostgresqlEnabled `json:"db.server.query_sample,omitempty" yaml:"db.server.query_sample,omitempty"`
	DbServerTopQuery    PostgresqlEnabled `json:"db.server.top_query,omitempty" yaml:"db.server.top_query,omitempty"`
}
type PostgresqlEnabled struct {
	Enabled bool `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}
type PostgresqlQuerySampleCollection struct {
	MaxRowsPerQuery int `json:"max_rows_per_query,omitempty" yaml:"max_rows_per_query,omitempty"`
}
type PostgresqlTopQueryCollection struct {
	TopNQuery              int    `json:"top_n_query,omitempty" yaml:"top_n_query,omitempty"`
	MaxRowsPerQuery        int    `json:"max_rows_per_query,omitempty" yaml:"max_rows_per_query,omitempty"`
	MaxExplainEachInterval int    `json:"max_explain_each_interval,omitempty" yaml:"max_explain_each_interval,omitempty"`
	QueryPlanCacheSize     int    `json:"query_plan_cache_size,omitempty" yaml:"query_plan_cache_size,omitempty"`
	QueryPlanCacheTTL      string `json:"query_plan_cache_ttl,omitempty" yaml:"query_plan_cache_ttl,omitempty"`
}
type PostgresqlConnectionPool struct {
	MaxIdleTime string `json:"max_idle_time,omitempty" yaml:"max_idle_time,omitempty"`
	MaxLifetime string `json:"max_lifetime,omitempty" yaml:"max_lifetime,omitempty"`
	MaxIdle     int    `json:"max_idle,omitempty" yaml:"max_idle,omitempty"`
	MaxOpen     int    `json:"max_open,omitempty" yaml:"max_open,omitempty"`
}
