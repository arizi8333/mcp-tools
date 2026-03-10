package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// ToolExecutionCounter tracks total executions by tool and status.
	ToolExecutionCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mcp_tool_execution_total",
		Help: "Total number of MCP tool executions",
	}, []string{"tool", "status"})

	// ToolExecutionDuration tracks the latency of tool calls.
	ToolExecutionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mcp_tool_execution_duration_seconds",
		Help:    "Duration of MCP tool executions",
		Buckets: prometheus.DefBuckets,
	}, []string{"tool"})

	// RequestCounter tracks incoming requests to the registry.
	RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mcp_registry_requests_total",
		Help: "Total requests received by the tool registry",
	}, []string{"tool"})
)

// StartMetricsServer starts a background HTTP server to export Prometheus metrics.
func StartMetricsServer(addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return server.ListenAndServe()
}
