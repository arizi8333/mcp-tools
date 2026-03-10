package metrics

import (
	"context"
	"mcp-go/internal/clients"

	"github.com/mark3labs/mcp-go/mcp"
)

type MetricsExplorer struct {
	client clients.PrometheusClient
}

func NewMetricsExplorer(c clients.PrometheusClient) *MetricsExplorer {
	return &MetricsExplorer{client: c}
}

func (t *MetricsExplorer) Name() string {
	return "metrics_query"
}

func (t *MetricsExplorer) Description() string {
	return "Executes a PromQL query against a Prometheus server (Internal Platform Adapter)."
}

func (t *MetricsExplorer) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("query", mcp.Required(), mcp.Description("PromQL query string")),
		mcp.WithString("endpoint", mcp.Description("Prometheus endpoint")),
	}
}

func (t *MetricsExplorer) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := request.RequireString("query")

	result, err := t.client.Query(ctx, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	return mcp.NewToolResultText(result), nil
}
