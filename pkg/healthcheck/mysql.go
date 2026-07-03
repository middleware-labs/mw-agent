package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlReceiver struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func (r *MysqlReceiver) Validate() error {
	if r.Endpoint == "" {
		return fmt.Errorf("mysql: endpoint not configured")
	}
	if r.Username == "" {
		return fmt.Errorf("mysql: username not configured")
	}
	return nil
}

func (r *MysqlReceiver) CheckHealth(ctx context.Context) error {
	checks, err := r.CheckHealthDetailed(ctx)
	if err != nil {
		return err
	}

	var missing []string
	for _, c := range checks {
		if !c.Granted {
			missing = append(missing, c.Name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("mysql: missing permissions: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (r *MysqlReceiver) CheckHealthDetailed(ctx context.Context) ([]PermissionCheck, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	host, port := parseEndpoint(r.Endpoint, "3306")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", r.Username, r.Password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("mysql: failed to open connection: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(10 * time.Second)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("mysql: connection failed: %w", err)
	}

	return r.checkPermissions(ctx, db), nil
}

func (r *MysqlReceiver) checkPermissions(ctx context.Context, db *sql.DB) []PermissionCheck {
	checks := []PermissionCheck{
		{Name: "CONNECTED", Granted: true},
		{Name: "AUTHENTICATED", Granted: true},
		{Name: "REPLICATION CLIENT"},
		{Name: "PROCESS"},
		{Name: "SELECT ON performance_schema"},
	}

	_, err := db.ExecContext(ctx, "SHOW REPLICA STATUS")
	if err != nil {
		_, err = db.ExecContext(ctx, "SHOW SLAVE STATUS")
	}
	checks[2].Granted = err == nil
	checks[2].Err = err

	_, err = db.ExecContext(ctx, "SHOW PROCESSLIST")
	checks[3].Granted = err == nil
	checks[3].Err = err

	_, err = db.ExecContext(ctx, "SELECT 1 FROM performance_schema.events_statements_summary_by_digest LIMIT 1")
	checks[4].Granted = err == nil
	checks[4].Err = err

	return checks
}
