package logs

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestLogAnalyzer_Execute(t *testing.T) {
	// Create a temporary log file
	tmpFile, err := os.CreateTemp("", "test_log_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := "line 1: info\nline 2: error: something went wrong\nline 3: debug\nline 4: error: another error"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	analyzer := &LogAnalyzer{}

	// Test Case: Find "error"
	// Note: We need a way to create a real mcp.CallToolRequest with params.
	// The library uses a BindArguments pattern or RequireString.
	// In our implementation, we use RequireString.

	// However, RequireString internally lookups in r.Params.Arguments.
	// We can simulate this by creating a request manually.
	req := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"path":    tmpFile.Name(),
				"pattern": "error",
			},
		},
	}

	result, err := analyzer.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.IsError {
		t.Fatalf("Tool returned error: %v", result.Content)
	}

	text := result.Content[0].(mcp.TextContent).Text
	if !strings.Contains(text, "Line 2") || !strings.Contains(text, "Line 4") {
		t.Errorf("Expected matches not found in result: %s", text)
	}
}
