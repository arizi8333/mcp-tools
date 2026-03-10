package logs

import (
	"context"
	"fmt"
	"mcp-go/internal/clients"
	"os"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

type LogAnalyzer struct {
	ESClient clients.ElasticsearchClient
}

func (t *LogAnalyzer) Name() string {
	return "logs_analyze"
}

func (t *LogAnalyzer) Description() string {
	return "Searches for specific patterns, errors, or summaries in log files or Elasticsearch. Supports multiple indices via wildcards or comma-separated values."
}

func (t *LogAnalyzer) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("source", mcp.Description("Log source: 'file' or 'elasticsearch' (default: auto-detect)")),
		mcp.WithString("path", mcp.Description("Path to the log file (for file source)")),
		mcp.WithString("index", mcp.Description("Elasticsearch index pattern. Supports: single (logstash-2024), wildcard (logs-*), or comma-separated (logs-app,logs-api)")),
		mcp.WithString("pattern", mcp.Required(), mcp.Description("Pattern or keyword to search for (supports wildcards)")),
		mcp.WithNumber("max_results", mcp.Description("Limit the number of results returned (default: 50)")),
		mcp.WithString("time_range", mcp.Description("Time range for Elasticsearch (e.g., '1h', '24h', '7d', default: '24h')")),
	}
}

func (t *LogAnalyzer) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pattern, err := request.RequireString("pattern")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("missing required parameter 'pattern': %v", err)), nil
	}

	maxResults := request.GetInt("max_results", 50)
	source := request.GetString("source", "")

	// Auto-detect source if not specified
	if source == "" {
		if request.GetString("index", "") != "" {
			source = "elasticsearch"
		} else if request.GetString("path", "") != "" {
			source = "file"
		} else {
			return mcp.NewToolResultError("either 'path' (for file) or 'index' (for elasticsearch) must be specified"), nil
		}
	}

	if source == "elasticsearch" {
		return t.searchElasticsearch(ctx, request, pattern, maxResults)
	}

	return t.searchFile(ctx, request, pattern, maxResults)
}

func (t *LogAnalyzer) searchFile(ctx context.Context, request mcp.CallToolRequest, pattern string, maxResults int) (*mcp.CallToolResult, error) {
	path := request.GetString("path", "")
	if path == "" {
		return mcp.NewToolResultError("'path' is required for file source"), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to read log file: %v", err)), nil
	}

	lines := strings.Split(string(data), "\n")
	var matches []string
	count := 0
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			matches = append(matches, fmt.Sprintf("Line %d: %s", i+1, line))
			count++
			if count >= maxResults {
				break
			}
		}
	}

	if len(matches) == 0 {
		return mcp.NewToolResultText("no matches found"), nil
	}

	return mcp.NewToolResultText(strings.Join(matches, "\n")), nil
}

func (t *LogAnalyzer) searchElasticsearch(ctx context.Context, request mcp.CallToolRequest, pattern string, maxResults int) (*mcp.CallToolResult, error) {
	if t.ESClient == nil {
		return mcp.NewToolResultError("Elasticsearch client not configured"), nil
	}

	// Elasticsearch natively supports:
	// - Wildcards: "logs-*"
	// - Comma-separated: "logs-app,logs-api,audit-*"
	// - Single index: "logstash-2024"
	index := request.GetString("index", "logs-*")
	timeRange := request.GetString("time_range", "24h")

	// Build Elasticsearch query
	query := map[string]interface{}{
		"size": maxResults,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"query_string": map[string]interface{}{
							"query": fmt.Sprintf("*%s*", pattern),
						},
					},
					{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": fmt.Sprintf("now-%s", timeRange),
							},
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{"@timestamp": map[string]string{"order": "desc"}},
		},
	}

	logs, err := t.ESClient.Search(ctx, index, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("elasticsearch search failed: %v", err)), nil
	}

	if len(logs) == 0 {
		return mcp.NewToolResultText("no matches found"), nil
	}

	return mcp.NewToolResultText(strings.Join(logs, "\n")), nil
}
