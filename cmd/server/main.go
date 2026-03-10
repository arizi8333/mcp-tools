package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mcp-go/internal"
	"mcp-go/internal/clients"
	"mcp-go/internal/config"
	"mcp-go/internal/server"
	"mcp-go/internal/tools/database"
	"mcp-go/internal/tools/incident"
	"mcp-go/internal/tools/loadtest"
	"mcp-go/internal/tools/logs"
	"mcp-go/internal/tools/metrics"
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()

	// 2. Initialize clients (Real if configured, otherwise Mock)
	var promClient clients.PrometheusClient
	if cfg.PrometheusAddr != "" {
		promClient, _ = clients.NewRealPrometheus(cfg.PrometheusAddr)
	} else {
		promClient = &clients.MockPrometheus{Endpoint: "http://localhost:9090"}
	}

	var esClient clients.ElasticsearchClient
	if cfg.ElasticsearchURL != "" {
		var err error
		esClient, err = clients.NewRealElasticsearch(cfg.ElasticsearchURL, cfg.ElasticsearchUser, cfg.ElasticsearchPass, cfg.ElasticsearchAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize Elasticsearch: %v\n", err)
			esClient = &clients.MockElasticsearch{BaseURL: cfg.ElasticsearchURL}
			fmt.Fprintf(os.Stderr, "Initialized Elasticsearch: MOCK (fallback due to error)\n")
		} else {
			fmt.Fprintf(os.Stderr, "Initialized Elasticsearch: %s\n", cfg.ElasticsearchURL)
		}
	} else {
		esClient = &clients.MockElasticsearch{BaseURL: "http://localhost:9200"}
		fmt.Fprintf(os.Stderr, "Initialized Elasticsearch: MOCK (no URL configured)\n")
	}

	dbClients := make(map[string]clients.DatabaseClient)

	// Priority 1: Load from config file if available
	if cfg.DataSourcesConfig != nil && len(cfg.DataSourcesConfig.Databases) > 0 {
		for _, dbCfg := range cfg.DataSourcesConfig.Databases {
			if !dbCfg.Enabled {
				continue
			}
			c, err := clients.NewRealDatabase(dbCfg.Driver, dbCfg.DSN)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to initialize database '%s': %v\n", dbCfg.Name, err)
				continue
			}
			dbClients[dbCfg.Name] = c
			fmt.Fprintf(os.Stderr, "Initialized database: %s (%s)\n", dbCfg.Name, dbCfg.Driver)
		}
	} else {
		// Priority 2: Fallback to ENV-based configuration
		if cfg.DBDSN != "" {
			c, _ := clients.NewRealDatabase(cfg.DBDriver, cfg.DBDSN)
			dbClients["default"] = c
			dbClients[cfg.DBDriver] = c
			fmt.Fprintf(os.Stderr, "Initialized database: default (%s)\n", cfg.DBDriver)
		}
		if cfg.PostgresDSN != "" {
			c, _ := clients.NewRealDatabase("postgres", cfg.PostgresDSN)
			dbClients["postgres"] = c
			fmt.Fprintf(os.Stderr, "Initialized database: postgres\n")
		}
		if cfg.MySQLDSN != "" {
			c, _ := clients.NewRealDatabase("mysql", cfg.MySQLDSN)
			dbClients["mysql"] = c
			fmt.Fprintf(os.Stderr, "Initialized database: mysql\n")
		}
	}

	// Fallback to mock if no databases configured
	if len(dbClients) == 0 {
		dbClients["default"] = &clients.MockDatabase{Driver: "postgres", DSN: "omitted"}
		fmt.Fprintf(os.Stderr, "Initialized database: default (MOCK)\n")
	}

	// 3. Initialize Platform Service
	service, err := internal.NewPlatformService(cfg, promClient, dbClients)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize platform service: %v\n", err)
		os.Exit(1)
	}

	// 4. Register Platform Capabilities
	service.RegisterResources()
	service.RegisterPrompts()
	service.RegisterTools(
		&internal.HealthTool{StartTime: time.Now()},
		&logs.LogAnalyzer{ESClient: esClient},
		metrics.NewMetricsExplorer(promClient),
		&loadtest.LoadTestAssistant{},
		database.NewDBInspector(service.DbClients),
		database.NewDBSchemaTool(service.DbClients),
		database.NewDBQueryTool(service.DbClients),
		&incident.IncidentAnalyzer{},
	)

	// 5. Handle OS Signals for Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// 5. Start Server (Stdio or HTTP)
	done := make(chan error, 1)

	if cfg.Transport == "http" {
		// HTTP Mode (SSE)
		go func() {
			// Use PUBLIC_URL if set, otherwise construct from HOST:PORT
			publicURL := cfg.PublicURL
			if publicURL == "" {
				publicURL = fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)
			}
			fmt.Printf("Starting MCP server in HTTP mode on %s\n", publicURL)
			done <- server.StartHTTPServer(publicURL, service.MCPServer)
		}()
	} else {
		// Stdio Mode (Default)
		go func() {
			done <- service.Start()
		}()
	}

	select {
	case err := <-done:
		if err != nil {
			fmt.Fprintf(os.Stderr, "Platform Service fatal error: %v\n", err)
			os.Exit(1)
		}
	case sig := <-stop:
		fmt.Fprintf(os.Stderr, "Received signal %v: shutting down gracefully...\n", sig)
		os.Exit(0)
	}
}
