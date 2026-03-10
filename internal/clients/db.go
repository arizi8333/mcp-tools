package clients

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseClient defines the interface for database operations.
type DatabaseClient interface {
	Explain(ctx context.Context, query string) (string, error)
	GetSchema(ctx context.Context) (string, error)
	Query(ctx context.Context, query string) (string, error)
	Exec(ctx context.Context, query string) (string, error)
}

// RealDatabase is the production client wrapper using standard sql.DB.
type RealDatabase struct {
	db     *sql.DB
	driver string
}

// NewRealDatabase initializes a real database connection.
func NewRealDatabase(driver, dsn string) (DatabaseClient, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return &RealDatabase{db: db, driver: driver}, nil
}

func (d *RealDatabase) Explain(ctx context.Context, query string) (string, error) {
	explainQuery := fmt.Sprintf("EXPLAIN %s", query)
	rows, err := d.db.QueryContext(ctx, explainQuery)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var plan []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		plan = append(plan, line)
	}
	return fmt.Sprintf("Execution Plan:\n%v", plan), nil
}

func (d *RealDatabase) Query(ctx context.Context, query string) (string, error) {
	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return "", err
	}

	result := fmt.Sprintf("Query Results (%d columns):\n", len(cols))
	for i, col := range cols {
		result += col
		if i < len(cols)-1 {
			result += " | "
		}
	}
	result += "\n" + "---" + "\n"

	// Simplified row fetching for string output
	count := 0
	for rows.Next() {
		columns := make([]any, len(cols))
		columnPointers := make([]any, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return "", err
		}

		for i := range cols {
			val := columns[i]
			if b, ok := val.([]byte); ok {
				result += string(b)
			} else {
				result += fmt.Sprintf("%v", val)
			}
			if i < len(cols)-1 {
				result += " | "
			}
		}
		result += "\n"
		count++
		if count >= 100 { // Safety limit
			result += "... (truncated to 100 rows)"
			break
		}
	}

	if count == 0 {
		return "Query executed successfully. Result set is empty.", nil
	}

	return result, nil
}

func (d *RealDatabase) GetSchema(ctx context.Context) (string, error) {
	var query string
	switch d.driver {
	case "postgres":
		query = `
			SELECT table_name, column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = 'public'
			ORDER BY table_name, ordinal_position`
	case "mysql":
		query = `
			SELECT table_name, column_name, data_type 
			FROM information_schema.columns 
			WHERE table_schema = DATABASE()
			ORDER BY table_name, ordinal_position`
	default:
		return "", fmt.Errorf("schema inspection not supported for driver: %s", d.driver)
	}

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	schema := "Database Schema:\n"
	for rows.Next() {
		var table, column, dtype string
		if err := rows.Scan(&table, &column, &dtype); err != nil {
			return "", err
		}
		schema += fmt.Sprintf("- Table: %s | Column: %s (%s)\n", table, column, dtype)
	}
	return schema, nil
}

func (d *RealDatabase) Exec(ctx context.Context, query string) (string, error) {
	result, err := d.db.ExecContext(ctx, query)
	if err != nil {
		return "", err
	}
	rows, _ := result.RowsAffected()
	return fmt.Sprintf("Command executed successfully. Rows affected: %d", rows), nil
}

// MockDatabase is a simulation client for database operations.
type MockDatabase struct {
	Driver string
	DSN    string
}

func (m *MockDatabase) Explain(ctx context.Context, query string) (string, error) {
	return fmt.Sprintf("[Mock Explain] Driver: %s | Query: %s\nPlan: Sequential scan on large_table expected.", m.Driver, query), nil
}

func (m *MockDatabase) GetSchema(ctx context.Context) (string, error) {
	return "Database Schema (Mock):\n- Table: users | Column: id (int), email (varchar)\n- Table: logs | Column: id (int), level (varchar), message (text)", nil
}

func (m *MockDatabase) Query(ctx context.Context, query string) (string, error) {
	return fmt.Sprintf("[Mock Query] Executing: %s\nResults: Mock data returned successfully.", query), nil
}

func (m *MockDatabase) Exec(ctx context.Context, query string) (string, error) {
	return fmt.Sprintf("[Mock Exec] Executing: %s\nStatus: Success", query), nil
}
