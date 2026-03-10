package incident

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// IncidentAnalyzer is a meta-tool that provides unified diagnostic views.
type IncidentAnalyzer struct{}

func (t *IncidentAnalyzer) Name() string {
	return "incident_analyze"
}

func (t *IncidentAnalyzer) Description() string {
	return "Orchestrates logs, metrics, and database tools to provide a comprehensive incident diagnosis (Meta-Tool)."
}

func (t *IncidentAnalyzer) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("service", mcp.Required(), mcp.Description("The service name experiencing the incident")),
		mcp.WithString("time_range", mcp.Description("Time range to analyze (e.g., 'last 5 minutes')")),
	}
}

func (t *IncidentAnalyzer) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	service, _ := request.RequireString("service")

	// In this simulation, we orchestrate cross-domain data:
	diagnosis := []string{
		fmt.Sprintf("--- Incident Diagnosis for %s ---", service),
		"1. [Metrics] Error rate spiked to 12% in the last 5 minutes.",
		"2. [Logs] Multiple 'connection pool exhausted' errors found.",
		"3. [Database] 3 slow queries detected involving the 'orders' table.",
		"\nRecommendation: Increase database connection pool size and check index on 'orders' table.",
	}

	return mcp.NewToolResultText(strings.Join(diagnosis, "\n")), nil
}
