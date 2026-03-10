package internal

import (
	"context"
	"fmt"
	"time"

	"mcp-go/internal/audit"
	"mcp-go/internal/auth"
	"mcp-go/internal/cache"
	"mcp-go/internal/clients"
	"mcp-go/internal/config"
	"mcp-go/internal/executor"
	"mcp-go/internal/middleware"
	"mcp-go/internal/observability"
	"mcp-go/internal/registry"
	"mcp-go/internal/transport"
	"mcp-go/pkg/mcp"

	pkgmcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// PlatformService is the main orchestrator for the MCP internal platform.
type PlatformService struct {
	Config      *config.Config
	MCPServer   *server.MCPServer
	Registry    *registry.Registry
	Executor    *executor.Engine
	Auth        *auth.Authenticator
	Obs         *observability.Logger
	Audit       *audit.Store
	Transport   transport.Server
	PromClient  clients.PrometheusClient
	DbClients   map[string]clients.DatabaseClient
	RateLimiter *middleware.RateLimiter
	Cache       *cache.Cache
}

func NewPlatformService(cfg *config.Config, prom clients.PrometheusClient, dbs map[string]clients.DatabaseClient) (*PlatformService, error) {
	// Initialize Structured Logging (Zap)
	zapLog, err := observability.NewZapLogger("info")
	if err != nil {
		return nil, fmt.Errorf("failed to init zap logger: %w", err)
	}

	// Initialize Audit Store (SQLite)
	auditStore, err := audit.NewStore("audit_trail.db")
	if err != nil {
		return nil, fmt.Errorf("failed to init audit store: %w", err)
	}

	obs := &observability.Logger{
		Zap:   zapLog,
		Audit: auditStore,
	}

	exec := executor.NewEngine(10, 30*time.Second) // Defaults
	rl := middleware.NewRateLimiter(5, 10)         // 5 req/s, burst 10
	c := cache.NewCache()
	a := auth.NewAuthenticator()
	reg := registry.NewRegistry(exec, obs, rl, c, a)

	s := server.NewMCPServer(cfg.ServerName, cfg.ServerVersion)

	return &PlatformService{
		Config:      cfg,
		MCPServer:   s,
		Registry:    reg,
		Executor:    exec,
		Auth:        a,
		Obs:         obs,
		Audit:       auditStore,
		Transport:   &transport.StdioServer{},
		PromClient:  prom,
		DbClients:   dbs,
		RateLimiter: rl,
		Cache:       c,
	}, nil
}

func (p *PlatformService) RegisterTools(tools ...mcp.Tool) {
	p.Registry.Register(p.MCPServer, tools...)
}

func (p *PlatformService) RegisterResources() {
	// Platform Architecture Resource
	res := pkgmcp.NewResource("platform://architecture", "Platform Architecture Overview",
		pkgmcp.WithResourceDescription("Returns documentation on the layered platform architecture (Executor, Registry, etc.)"))

	p.MCPServer.AddResource(res, func(ctx context.Context, request pkgmcp.ReadResourceRequest) ([]pkgmcp.ResourceContents, error) {
		return []pkgmcp.ResourceContents{
			pkgmcp.TextResourceContents{
				URI:      "platform://architecture",
				MIMEType: "text/plain",
				Text:     "The MCP Internal Platform is a layered service specializing in reliable tool execution and system orchestration.",
			},
		}, nil
	})
}

func (p *PlatformService) RegisterPrompts() {
	// Incident Triage Prompt
	prompt := pkgmcp.NewPrompt("incident-triage",
		pkgmcp.WithPromptDescription("Standard template for triaging service incidents."),
		pkgmcp.WithArgument("service", pkgmcp.ArgumentDescription("Target service name"), pkgmcp.RequiredArgument()),
	)

	p.MCPServer.AddPrompt(prompt, func(ctx context.Context, request pkgmcp.GetPromptRequest) (*pkgmcp.GetPromptResult, error) {
		service := request.Params.Arguments["service"]
		return &pkgmcp.GetPromptResult{
			Description: "Triage Template",
			Messages: []pkgmcp.PromptMessage{
				{
					Role: pkgmcp.RoleUser,
					Content: pkgmcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("I am analyzing an incident for '%s'. Please use the platform tools to diagnose metrics and logs.", service),
					},
				},
			},
		}, nil
	})

	// Auto-Fixer Prompt
	fixer := pkgmcp.NewPrompt("auto-fixer",
		pkgmcp.WithPromptDescription("AI-assisted system repair and diagnosis guide."),
		pkgmcp.WithArgument("issue", pkgmcp.ArgumentDescription("Description of the issue to fix"), pkgmcp.RequiredArgument()),
	)

	p.MCPServer.AddPrompt(fixer, func(ctx context.Context, request pkgmcp.GetPromptRequest) (*pkgmcp.GetPromptResult, error) {
		issue := request.Params.Arguments["issue"]
		return &pkgmcp.GetPromptResult{
			Description: "Repair Strategy Template",
			Messages: []pkgmcp.PromptMessage{
				{
					Role: pkgmcp.RoleUser,
					Content: pkgmcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("I need to fix: %s. Please follow these steps:\n1. Check metrics.query\n2. Analyze logs.analyze\n3. Inspect db.schema and db.inspect\n4. Suggest command-line fixes.", issue),
					},
				},
			},
		}, nil
	})
}

func (p *PlatformService) Start() error {
	p.Obs.Info("Starting MCP Internal Platform Service: %s@%s", p.Config.ServerName, p.Config.ServerVersion)

	// Start Prometheus Metrics Server in the background
	go func() {
		p.Obs.Info("Starting internal metrics server on :9090")
		if err := observability.StartMetricsServer(":9090"); err != nil {
			p.Obs.Error("Failed to start metrics server", err)
		}
	}()

	return p.Transport.Serve(p.MCPServer)
}
