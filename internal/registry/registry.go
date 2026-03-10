package registry

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"mcp-go/internal/auth"
	"mcp-go/internal/cache"
	"mcp-go/internal/executor"
	"mcp-go/internal/middleware"
	"mcp-go/internal/observability"
	"mcp-go/pkg/mcp"

	pkgmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Registry struct {
	mu       sync.RWMutex
	tools    map[string]mcp.Tool
	executor *executor.Engine
	obs      *observability.Logger
	limiter  *middleware.RateLimiter
	cache    *cache.Cache
	auth     *auth.Authenticator
}

func NewRegistry(e *executor.Engine, obs *observability.Logger, rl *middleware.RateLimiter, c *cache.Cache, a *auth.Authenticator) *Registry {
	return &Registry{
		tools:    make(map[string]mcp.Tool),
		executor: e,
		obs:      obs,
		limiter:  rl,
		cache:    c,
		auth:     a,
	}
}

func (r *Registry) Register(s *server.MCPServer, tools ...mcp.Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, t := range tools {
		r.tools[t.Name()] = t

		// Register tool with parameters from implementation
		opts := []pkgmcp.ToolOption{
			pkgmcp.WithDescription(t.Description()),
		}
		opts = append(opts, t.Parameters()...)

		s.AddTool(pkgmcp.NewTool(t.Name(), opts...), r.handleExecution)
	}
}

func (r *Registry) handleExecution(ctx context.Context, request pkgmcp.CallToolRequest) (*pkgmcp.CallToolResult, error) {
	r.mu.RLock()
	tool, ok := r.tools[request.Params.Name]
	r.mu.RUnlock()

	if !ok {
		return pkgmcp.NewToolResultError(fmt.Sprintf("tool not found: %s", request.Params.Name)), nil
	}

	// 1. Authentication & RBAC Check
	apiKey := request.GetString("_api_key", "") // Use special param for auth in stdio
	role, err := r.auth.Authenticate(ctx, apiKey)
	if err != nil {
		return pkgmcp.NewToolResultError(err.Error()), nil
	}

	if !r.auth.CheckPermission(role, tool.Name()) {
		return pkgmcp.NewToolResultError(fmt.Sprintf("unauthorized: role %s cannot access %s", role, tool.Name())), nil
	}

	// 2. Rate Limiting Protection
	if !r.limiter.Allow(tool.Name()) {
		return pkgmcp.NewToolResultError(fmt.Sprintf("rate limit exceeded for tool: %s", tool.Name())), nil
	}

	// 3. Execution Caching Layer
	cacheKey := r.generateCacheKey(request)
	if val, found := r.cache.Get(cacheKey); found {
		if res, ok := val.(*pkgmcp.CallToolResult); ok {
			r.obs.Info("Cache HIT for tool: %s", tool.Name())
			return res, nil
		}
	}

	r.obs.IncrementRequestCount(tool.Name())

	startTime := time.Now()
	res, err := r.executor.Execute(ctx, tool, request)
	r.obs.LogExecution(tool.Name(), time.Since(startTime), err == nil && !res.IsError)

	if err == nil && !res.IsError {
		// Cache successful results for 1 minute for idempotency
		r.cache.Set(cacheKey, res, 1*time.Minute)
	}

	return res, err
}

func (r *Registry) generateCacheKey(req pkgmcp.CallToolRequest) string {
	data, _ := json.Marshal(req.Params)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}
