package mcp

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

// Tool defines the standard interface for all platform tools.
type Tool interface {
	Name() string
	Description() string
	Parameters() []mcp.ToolOption
	Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// ExecutionOptions defines common constraints for tool execution.
type ExecutionOptions struct {
	TimeoutSeconds int
	MaxConcurrency int
}
