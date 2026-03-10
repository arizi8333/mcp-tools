package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/mark3labs/mcp-go/server"
)

// StartHTTPServer starts the MCP server in HTTP mode (Streamable HTTP / SSE)
func StartHTTPServer(publicURL string, mcpServer *server.MCPServer) error {
	// Parse public URL safely
	u, err := url.Parse(publicURL)
	if err != nil {
		return fmt.Errorf("invalid public URL: %w", err)
	}

	// Internal listen address
	listenAddr := ":8080"
	if port := u.Port(); port != "" {
		listenAddr = ":" + port
	}

	// Use StreamableHTTPServer in stateless mode (supports POST without sessionId)
	streamableHandler := server.NewStreamableHTTPServer(
		mcpServer,
		server.WithEndpointPath("/sse"),
		server.WithStateLess(true), // No sessionId required
	)

	mux := http.NewServeMux()

	// ---- Main MCP endpoint (stateless streamable HTTP) ----
	mux.Handle("/sse", withCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[DEBUG] %s %s - Query: %s, ContentType: %s", r.Method, r.URL.Path, r.URL.RawQuery, r.Header.Get("Content-Type"))
		streamableHandler.ServeHTTP(w, r)
	})))

	// Optional aliases
	mux.Handle("/mcp", withCORS(streamableHandler))
	mux.Handle("/message", withCORS(streamableHandler))
	mux.Handle("/messages", withCORS(streamableHandler))

	// ---- Health ----
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Listening on %s (Streamable HTTP mode - stateless)", listenAddr)

	return http.ListenAndServe(listenAddr, logRequests(mux))
}

// CORS middleware
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Request logger (very useful for MCP debugging)
func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}
