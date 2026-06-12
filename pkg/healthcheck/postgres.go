package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type PostgresqlReceiver struct {
	Endpoint string `mapstructure:"endpoint"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

func (r *PostgresqlReceiver) Validate() error {
	if r.Endpoint == "" {
		return fmt.Errorf("postgresql: endpoint not configured")
	}
	if r.Username == "" {
		return fmt.Errorf("postgresql: username not configured")
	}
	return nil
}

func (r *PostgresqlReceiver) CheckHealth(ctx context.Context) error {
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
		return fmt.Errorf("postgresql: missing permissions: %s", strings.Join(missing, ", "))
	}

	return nil
}

func (r *PostgresqlReceiver) CheckHealthDetailed(ctx context.Context) ([]PermissionCheck, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	host, port := parseEndpoint(r.Endpoint, "5432")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable connect_timeout=5",
		host, port, r.Username, r.Password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("postgresql: failed to open connection: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(10 * time.Second)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("postgresql: connection failed: %w", err)
	}

	return r.checkPermissions(ctx, db), nil
}

func (r *PostgresqlReceiver) checkPermissions(ctx context.Context, db *sql.DB) []PermissionCheck {
	checks := []PermissionCheck{
		{Name: "CONNECTED", Granted: true},
		{Name: "AUTHENTICATED", Granted: true},
		{Name: "pg_stat_statements"},
		{Name: "pg_read_all_stats"},
	}

	_, err := db.ExecContext(ctx, "SELECT 1 FROM pg_stat_statements LIMIT 1")
	checks[2].Granted = err == nil
	checks[2].Err = err

	var hasRole bool
	err = db.QueryRowContext(ctx, "SELECT pg_has_role(current_user, 'pg_read_all_stats', 'MEMBER')").Scan(&hasRole)
	checks[3].Granted = err == nil && hasRole
	checks[3].Err = err

	return checks
}
