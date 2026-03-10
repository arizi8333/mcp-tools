package transport

import (
	"fmt"

	"github.com/mark3labs/mcp-go/server"
)

// Server defines the interface for an MCP transport server.
type Server interface {
	Serve(s *server.MCPServer) error
}

// StdioServer implements MCP communication over standard input/output.
type StdioServer struct{}

func (s *StdioServer) Serve(mcpServer *server.MCPServer) error {
	if err := server.ServeStdio(mcpServer); err != nil {
		return fmt.Errorf("stdio server failed: %w", err)
	}
	return nil
}
