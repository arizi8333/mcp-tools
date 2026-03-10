package registry

import (
	"context"
	"mcp-go/internal/auth"
	"mcp-go/internal/cache"
	"mcp-go/internal/executor"
	"mcp-go/internal/middleware"
	"mcp-go/internal/observability"
	"testing"
	"time"

	pkgmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type mockTool struct{}

func (m *mockTool) Name() string                    { return "test.tool" }
func (m *mockTool) Description() string             { return "description" }
func (m *mockTool) Parameters() []pkgmcp.ToolOption { return nil }
func (m *mockTool) Execute(ctx context.Context, req pkgmcp.CallToolRequest) (*pkgmcp.CallToolResult, error) {
	return pkgmcp.NewToolResultText("ok"), nil
}

func TestRegistry_Register(t *testing.T) {
	obs := &observability.Logger{}
	exec := executor.NewEngine(5, 1*time.Second)
	rl := middleware.NewRateLimiter(10, 20)
	c := cache.NewCache()
	a := auth.NewAuthenticator()
	reg := NewRegistry(exec, obs, rl, c, a)
	s := server.NewMCPServer("test", "1.0")

	reg.Register(s, &mockTool{})

	if _, ok := reg.tools["test.tool"]; !ok {
		t.Error("tool was not registered in the registry map")
	}
}
