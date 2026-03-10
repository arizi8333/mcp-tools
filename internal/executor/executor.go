package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"mcp-go/pkg/mcp"

	pkgmcp "github.com/mark3labs/mcp-go/mcp"
)

// Engine manages the execution of MCP tools.
type Engine struct {
	mu             sync.Mutex
	activeWork     int
	maxConcurrency int
	defaultTimeout time.Duration
}

func NewEngine(maxConcurrency int, defaultTimeout time.Duration) *Engine {
	return &Engine{
		maxConcurrency: maxConcurrency,
		defaultTimeout: defaultTimeout,
	}
}

// Execute wraps a tool execution with safety controls.
func (e *Engine) Execute(ctx context.Context, tool mcp.Tool, request pkgmcp.CallToolRequest) (*pkgmcp.CallToolResult, error) {
	// 1. Concurrency Check
	e.mu.Lock()
	if e.maxConcurrency > 0 && e.activeWork >= e.maxConcurrency {
		e.mu.Unlock()
		return pkgmcp.NewToolResultError("server is busy: max concurrency reached"), nil
	}
	e.activeWork++
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.activeWork--
		e.mu.Unlock()
	}()

	// 2. Setup Timeout & Cancellation
	timeout := e.defaultTimeout
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if toolTimeout, ok := args["_timeout"].(float64); ok {
			timeout = time.Duration(toolTimeout) * time.Second
		}
	}

	executionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 3. Execution Channel
	resultChan := make(chan struct {
		result *pkgmcp.CallToolResult
		err    error
	}, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- struct {
					result *pkgmcp.CallToolResult
					err    error
				}{pkgmcp.NewToolResultError(fmt.Sprintf("tool execution panicked: %v", r)), nil}
			}
		}()

		res, err := tool.Execute(executionCtx, request)
		resultChan <- struct {
			result *pkgmcp.CallToolResult
			err    error
		}{res, err}
	}()

	// 4. Wait for Result or Timeout/Cancellation
	select {
	case res := <-resultChan:
		return res.result, res.err
	case <-executionCtx.Done():
		if executionCtx.Err() == context.DeadlineExceeded {
			return pkgmcp.NewToolResultError(fmt.Sprintf("tool execution timed out after %s", timeout)), nil
		}
		return pkgmcp.NewToolResultError("tool execution cancelled"), nil
	}
}
