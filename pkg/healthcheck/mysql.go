package healthcheck

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/k0kubun/pp"
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
	pp.Println("=== MySQL Health Check ===")
	pp.Println("Endpoint:", r.Endpoint)
	pp.Println("Username:", r.Username)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	host, port := parseEndpoint(r.Endpoint, "3306")
	pp.Println("Parsed host:", host, "port:", port)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/?timeout=5s", r.Username, r.Password, host, port)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		pp.Println("FAIL: could not open connection:", err)
		return fmt.Errorf("mysql: failed to open connection: %w", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(10 * time.Second)

	pp.Println("Pinging database...")
	if err := db.PingContext(ctx); err != nil {
		pp.Println("FAIL: ping failed:", err)
		return fmt.Errorf("mysql: connection failed: %w", err)
	}
	pp.Println("OK: connection established")

	pp.Println("Checking permissions...")
	checks := r.checkPermissions(ctx, db)

	for _, c := range checks {
		if c.Granted {
			pp.Println("  ✓", c.Name)
		} else {
			pp.Println("  ✗", c.Name, "—", c.Err)
		}
	}

	var missing []string
	for _, c := range checks {
		if !c.Granted {
			missing = append(missing, c.Name)
		}
	}
	if len(missing) > 0 {
		pp.Println("FAIL: missing permissions:", missing)
		return fmt.Errorf("mysql: missing permissions: %s", strings.Join(missing, ", "))
	}

	pp.Println("=== MySQL Health Check PASSED ===")
	return nil
}

func (r *MysqlReceiver) checkPermissions(ctx context.Context, db *sql.DB) []PermissionCheck {
	checks := []PermissionCheck{
		{Name: "CONNECT", Granted: true},
		{Name: "REPLICATION CLIENT"},
		{Name: "PROCESS"},
		{Name: "SELECT ON performance_schema"},
	}

	// REPLICATION CLIENT — SHOW REPLICA STATUS (8.0.22+) or SHOW SLAVE STATUS (older)
	pp.Println("Running: SHOW REPLICA STATUS")
	err := printQueryRows(ctx, db, "SHOW REPLICA STATUS")
	if err != nil {
		pp.Println("Falling back: SHOW SLAVE STATUS")
		err = printQueryRows(ctx, db, "SHOW SLAVE STATUS")
	}
	checks[1].Granted = err == nil
	checks[1].Err = err

	// PROCESS — SHOW PROCESSLIST requires it
	pp.Println("Running: SHOW PROCESSLIST")
	err = printQueryRows(ctx, db, "SHOW PROCESSLIST")
	checks[2].Granted = err == nil
	checks[2].Err = err

	// SELECT on performance_schema
	pp.Println("Running: SELECT 1 FROM performance_schema.events_statements_summary_by_digest LIMIT 1")
	err = printQueryRows(ctx, db, "SELECT 1 FROM performance_schema.events_statements_summary_by_digest LIMIT 1")
	checks[3].Granted = err == nil
	checks[3].Err = err

	return checks
}

func printQueryRows(ctx context.Context, db *sql.DB, query string) error {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		pp.Println("  err:", err)
		return err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	pp.Println("  columns:", cols)

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	rowNum := 0
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			pp.Println("  scan err:", err)
			continue
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = vals[i]
		}
		pp.Printf("  row[%d]: %v\n", rowNum, row)
		rowNum++
	}
	if rowNum == 0 {
		pp.Println("  (no rows)")
	}
	return rows.Err()
}
