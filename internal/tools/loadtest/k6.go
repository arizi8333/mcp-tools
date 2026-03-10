package loadtest

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/mark3labs/mcp-go/mcp"
)

type LoadTestAssistant struct{}

func (t *LoadTestAssistant) Name() string {
	return "loadtest_run"
}

func (t *LoadTestAssistant) Description() string {
	return "Executes a k6 load test script in a controlled sandbox."
}

func (t *LoadTestAssistant) Parameters() []mcp.ToolOption {
	return []mcp.ToolOption{
		mcp.WithString("script_path", mcp.Required(), mcp.Description("Path to the k6 script file")),
	}
}

func (t *LoadTestAssistant) Execute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scriptPath, _ := request.RequireString("script_path")

	// Check if k6 is installed
	_, err := exec.LookPath("k6")
	if err != nil {
		return mcp.NewToolResultError("k6 is not installed or not in PATH"), nil
	}

	// Important: Use ctx for cancellation/timeout support
	cmd := exec.CommandContext(ctx, "k6", "run", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError("k6 execution exceeded allotted tool timeout"), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("k6 execution failed: %v\nOutput: %s", err, string(output))), nil
	}

	return mcp.NewToolResultText(string(output)), nil
}
