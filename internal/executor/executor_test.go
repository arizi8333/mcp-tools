package executor

import (
	"context"
	"testing"
	"time"

	pkgmcp "github.com/mark3labs/mcp-go/mcp"
)

type mockTool struct {
	delay time.Duration
	err   error
}

func (m *mockTool) Name() string                    { return "mock" }
func (m *mockTool) Description() string             { return "mock tool" }
func (m *mockTool) Parameters() []pkgmcp.ToolOption { return nil }
func (m *mockTool) Execute(ctx context.Context, req pkgmcp.CallToolRequest) (*pkgmcp.CallToolResult, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if m.err != nil {
		return nil, m.err
	}
	return pkgmcp.NewToolResultText("ok"), nil
}

func TestEngine_Execute_Timeout(t *testing.T) {
	engine := NewEngine(5, 50*time.Millisecond)
	tool := &mockTool{delay: 100 * time.Millisecond}

	req := pkgmcp.CallToolRequest{
		Params: pkgmcp.CallToolParams{Name: "mock"},
	}

	res, err := engine.Execute(context.Background(), tool, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsError {
		t.Error("expected timeout error in result, but got success")
	}
}

func TestEngine_Execute_Concurrency(t *testing.T) {
	engine := NewEngine(1, 1*time.Second)
	tool := &mockTool{delay: 100 * time.Millisecond}

	req := pkgmcp.CallToolRequest{
		Params: pkgmcp.CallToolParams{Name: "mock"},
	}

	// Start first one
	go engine.Execute(context.Background(), tool, req)
	time.Sleep(10 * time.Millisecond)

	// Second one should hit concurrency limit
	res, err := engine.Execute(context.Background(), tool, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsError {
		t.Error("expected concurrency limit error, but got success")
	}
}
