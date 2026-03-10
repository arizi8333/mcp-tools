# MCP Testing & Implementation Playbook

This guide covers how to implement and test features in the MCP Internal Platform Service.

## 1. Automated Testing (Unit Tests)

The platform includes a comprehensive unit test suite covering the core execution logic and tool implementations.

### Running Tests
Execute the following command to run all tests:
```bash
make test
```

### Key Test Components
- **Executor Tests (`internal/executor/executor_test.go`)**: Validates that the engine correctly handles tool timeouts and blocks execution when concurrency limits are reached.
- **Registry Tests (`internal/registry/registry_test.go`)**: Ensures tools are correctly registered and mapped to the protocol handler.
- **Tool Tests (e.g., `internal/tools/logs/logs_test.go`)**: Validates the business logic of individual tools in isolation.

---

## 2. Manual Implementation & Integration

### Adding a New Capability
To add a new tool, follow this flow:
1.  **Define Interface**: Create a client adapter in `internal/clients/` if external systems are involved.
2.  **Implement Engine Logic**: Create a new tool in `internal/tools/` implementing `pkg/mcp.Tool`.
3.  **Bootstrapping**: Initialize the client and tool in `cmd/server/main.go` and register it with `service.RegisterTools()`.

---

## 3. Manual Verification (E2E)

Since MCP uses Stdio, you can verify tool execution by piping JSON-RPC messages directly to the server binary.

### Health Check Execution
To verify the service is alive:
1. Build the binary: `make build`
2. Send a JSON-RPC request:
```bash
echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "platform.health", "arguments": {}}, "id": 1}' | ./mcp-server
```

### Timeout Verification
You can pass a `_timeout` argument to any tool (supported by the Platform Executor) to force a timeout:
```bash
echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "logs.analyze", "arguments": {"path": "test.log", "pattern": "error", "_timeout": 0.001}}, "id": 2}' | ./mcp-server
```

### Metrics Verification
The platform automatically starts a Prometheus metrics server on `:9090`. You can verify metrics are being collected by visiting:
```bash
curl http://localhost:9090/metrics
```
Look for `mcp_tool_execution_total` and `mcp_tool_execution_duration_seconds`.

### Rate Limiting & Cache Verification
You can observe Rate Limiting and Cache behavior in the server logs when running in debug/info mode. Repeated calls to the same tool will show "Cache HIT" messages in the audit trail.

### Multi-Database Verification
If multiple databases are configured (e.g., Postgres and MySQL), you can verify them individually:
1.  **Schema Check (Postgres)**:
    ```bash
    echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "db.schema", "arguments": {"database": "postgres"}}, "id": 10}' | ./mcp-server
    ```
2.  **Schema Check (MySQL)**:
    ```bash
    echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "db.schema", "arguments": {"database": "mysql"}}, "id": 11}' | ./mcp-server
    ```

---

## 4. Integration with Claude Desktop
The most reliable way to test "feel" and "discovery" is by integrating with a real MCP client:
1. Add the server to `claude_desktop_config.json`.
2. Restart Claude Desktop.
3. Check the "Logs" in Claude Desktop to see the orchestrator (internal/platform.go) starting up and tool audit logs.
