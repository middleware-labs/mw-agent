package agent

import (
	"fmt"
)

// ========== Postgresql Receiver Config Updates ==========
func (postgresql *Postgresql) UpdateReceiverConfig(receiverConfig map[string]interface{}) error {
	postresqlReceiverConfig, ok := receiverConfig["postgresql"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("postgresql receiver config not found in receiverConfig")
	}
	// username (required)
	if postgresql.Username == "" {
		return fmt.Errorf("username is required field in postgresql integration configuration")
	}
	postresqlReceiverConfig["username"] = postgresql.Username
	// password (required)
	if postgresql.Password == "" {
		return fmt.Errorf("password is required field in postgresql integration configuration")
	}
	postresqlReceiverConfig["password"] = postgresql.Password
	// endpoint
	if postgresql.Endpoint != "" {
		postresqlReceiverConfig["endpoint"] = postgresql.Endpoint
	}
	// databases
	if len(postgresql.Databases) > 0 {
		postresqlReceiverConfig["databases"] = postgresql.Databases
	}
	// exclude_databases
	if len(postgresql.ExcludeDatabases) > 0 {
		postresqlReceiverConfig["exclude_databases"] = postgresql.ExcludeDatabases
	}
	// collection_interval
	if postgresql.CollectionInterval != "" {
		postresqlReceiverConfig["collection_interval"] = postgresql.CollectionInterval
	}
	// transport
	if postgresql.Transport != "" {
		postresqlReceiverConfig["transport"] = postgresql.Transport
	}
	// tls
	if postgresql.TLS != nil {
		tls := map[string]interface{}{
			"insecure":             postgresql.TLS.Insecure,
			"insecure_skip_verify": postgresql.TLS.InsecureSkipVerify,
		}
		if !postgresql.TLS.InsecureSkipVerify {
			if (postgresql.TLS.CertFile != "" && postgresql.TLS.KeyFile == "") || (postgresql.TLS.KeyFile != "" && postgresql.TLS.CertFile == "") {
				return fmt.Errorf("both cert_file and key_file must be provided together in TLS configuration for postgresql integration")
			} else if postgresql.TLS.CertFile != "" && postgresql.TLS.KeyFile != "" {
				tls["cert_file"] = postgresql.TLS.CertFile
				tls["key_file"] = postgresql.TLS.KeyFile
			}
			if postgresql.TLS.CaFile != "" {
				tls["ca_file"] = postgresql.TLS.CaFile
			}
		}
		postresqlReceiverConfig["tls"] = tls
	}
	// events
	if postgresql.Events != nil {
		postresqlReceiverConfig["events"] = map[string]interface{}{
			"db.server.query_sample": map[string]interface{}{
				"enabled": postgresql.Events.DbServerQuerySample.Enabled,
			},
			"db.server.top_query": map[string]interface{}{
				"enabled": postgresql.Events.DbServerTopQuery.Enabled,
			},
		}
	}
	// query_sample_collection
	if postgresql.QuerySampleCollection != nil {
		if postgresql.QuerySampleCollection.MaxRowsPerQuery <= 0 {
			postgresql.QuerySampleCollection.MaxRowsPerQuery = 1000
		}
		postresqlReceiverConfig["query_sample_collection"] = map[string]interface{}{
			"max_rows_per_query": postgresql.QuerySampleCollection.MaxRowsPerQuery,
		}
	}
	// top_query_collection
	if postgresql.TopQueryCollection != nil {
		if postgresql.TopQueryCollection.TopNQuery <= 0 {
			postgresql.TopQueryCollection.TopNQuery = 1000
		}
		if postgresql.TopQueryCollection.MaxRowsPerQuery <= 0 {
			postgresql.TopQueryCollection.MaxRowsPerQuery = 1000
		}
		if postgresql.TopQueryCollection.MaxExplainEachInterval <= 0 {
			postgresql.TopQueryCollection.MaxExplainEachInterval = 1000
		}
		if postgresql.TopQueryCollection.QueryPlanCacheSize <= 0 {
			postgresql.TopQueryCollection.QueryPlanCacheSize = 1000
		}
		if postgresql.TopQueryCollection.QueryPlanCacheTTL == "" {
			postgresql.TopQueryCollection.QueryPlanCacheTTL = "1h"
		}
		postresqlReceiverConfig["top_query_collection"] = map[string]interface{}{
			"top_n_query":               postgresql.TopQueryCollection.TopNQuery,
			"max_rows_per_query":        postgresql.TopQueryCollection.MaxRowsPerQuery,
			"max_explain_each_interval": postgresql.TopQueryCollection.MaxExplainEachInterval,
			"query_plan_cache_size":     postgresql.TopQueryCollection.QueryPlanCacheSize,
			"query_plan_cache_ttl":      postgresql.TopQueryCollection.QueryPlanCacheTTL,
		}
	}
	// connection_pool
	if postgresql.ConectionPool != nil {
		if postgresql.ConectionPool.MaxIdleTime == "" {
			postgresql.ConectionPool.MaxIdleTime = "10m"
		}
		if postgresql.ConectionPool.MaxLifetime == "" {
			postgresql.ConectionPool.MaxLifetime = "10m"
		}
		if postgresql.ConectionPool.MaxIdle <= 0 {
			postgresql.ConectionPool.MaxIdle = 10
		}
		if postgresql.ConectionPool.MaxOpen <= 0 {
			postgresql.ConectionPool.MaxOpen = 10
		}
		postresqlReceiverConfig["connection_pool"] = map[string]interface{}{
			"max_idle_time": postgresql.ConectionPool.MaxIdleTime,
			"max_lifetime":  postgresql.ConectionPool.MaxLifetime,
			"max_idle":      postgresql.ConectionPool.MaxIdle,
			"max_open":      postgresql.ConectionPool.MaxOpen,
		}
	}
	return nil
}
