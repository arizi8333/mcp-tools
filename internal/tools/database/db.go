package database

import (
	"context"
	"fmt"
	"mcp-go/internal/clients"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// DBInspector handles query explanation across multiple databases.
type DBInspector struct {
	clients map[string]clients.DatabaseClient
}

func NewDBInspector(cs map[string]clients.DatabaseClient) *DBInspector {
	return &DBInspector{clients: cs}
}

func (t *DBInspector) Name() string {
	return "db_inspect"
}

func (t *DBInspector) Description() string {
	return "Executes a read-only SQL query against a targeted database with EXPLAIN support."
}

func (t *DBInspector) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("query", mcp.Required(), mcp.Description("SQL query to explain")),
		mcp.WithString("database", mcp.Description("Target database name (if multiple are configured)")),
	}
}

func (t *DBInspector) getClient(request mcp.CallToolRequest) (clients.DatabaseClient, error) {
	dbName := request.GetString("database", "")
	if dbName == "" {
		if c, ok := t.clients["default"]; ok {
			return c, nil
		}
		if len(t.clients) == 1 {
			for _, v := range t.clients {
				return v, nil
			}
		}
		return nil, fmt.Errorf("database name is required when multiple databases are configured")
	}

	c, ok := t.clients[dbName]
	if !ok {
		return nil, fmt.Errorf("database '%s' not found", dbName)
	}
	return c, nil
}

func (t *DBInspector) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := request.RequireString("query")

	client, err := t.getClient(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := client.Explain(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

// DBSchemaTool handles database schema discovery across multiple databases.
type DBSchemaTool struct {
	clients map[string]clients.DatabaseClient
}

func NewDBSchemaTool(cs map[string]clients.DatabaseClient) *DBSchemaTool {
	return &DBSchemaTool{clients: cs}
}

func (t *DBSchemaTool) Name() string {
	return "db_schema"
}

func (t *DBSchemaTool) Description() string {
	return "Retrieves the database schema for a targeted database."
}

func (t *DBSchemaTool) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("database", mcp.Description("Target database name (if multiple are configured)")),
	}
}

func (t *DBSchemaTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName := request.GetString("database", "")

	// Re-using the same logic for simplicity
	var client clients.DatabaseClient
	if dbName == "" {
		if c, ok := t.clients["default"]; ok {
			client = c
		} else if len(t.clients) == 1 {
			for _, v := range t.clients {
				client = v
				break
			}
		} else {
			return mcp.NewToolResultError("database name is required when multiple databases are configured"), nil
		}
	} else {
		var ok bool
		client, ok = t.clients[dbName]
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("database '%s' not found", dbName)), nil
		}
	}

	result, err := client.GetSchema(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}

// DBQueryTool handles direct SQL query execution.
type DBQueryTool struct {
	clients map[string]clients.DatabaseClient
}

func NewDBQueryTool(cs map[string]clients.DatabaseClient) *DBQueryTool {
	return &DBQueryTool{clients: cs}
}

func (t *DBQueryTool) Name() string {
	return "db_query"
}

func (t *DBQueryTool) Description() string {
	return "Executes a SQL query against a targeted database and returns the result set."
}

func (t *DBQueryTool) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("query", mcp.Required(), mcp.Description("SQL query to execute")),
		mcp.WithString("database", mcp.Description("Target database name")),
	}
}

func (t *DBQueryTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := request.RequireString("query")

	// Reuse the same selection logic via a temporary inspector
	inspector := NewDBInspector(t.clients)
	client, err := inspector.getClient(request)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := "", error(nil)
	lowerQuery := strings.ToLower(strings.TrimSpace(query))
	if strings.HasPrefix(lowerQuery, "select") || strings.HasPrefix(lowerQuery, "with") || strings.HasPrefix(lowerQuery, "explain") {
		result, err = client.Query(ctx, query)
	} else {
		result, err = client.Exec(ctx, query)
	}

	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}
