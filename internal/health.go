package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

type HealthTool struct {
	StartTime time.Time
}

func (t *HealthTool) Name() string {
	return "platform_health"
}

func (t *HealthTool) Description() string {
	return "Returns the current health status and uptime of the MCP platform server."
}

func (t *HealthTool) Parameters() []mcp.ToolOption {
	return nil
}

func (t *HealthTool) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	uptime := time.Since(t.StartTime).String()
	return mcp.NewToolResultText(fmt.Sprintf("Service: Online\nUptime: %s\nStatus: Healthy", uptime)), nil
}
