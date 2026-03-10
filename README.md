# MCP Go Platform

A production-ready [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) server implementation in Go. This server acts as a centralized **Internal Platform Adapter**, connecting AI agents (Claude, OpenHands, etc.) to your internal infrastructure: databases, logs (Elasticsearch), metrics (Prometheus), and more.

## 🚀 Features

- **Protocol Support**:
  - **Stdio**: Standard JSON-RPC over stdin/stdout (default).
  - **HTTP (SSE)**: Server-Sent Events over HTTP for remote agents.
- **Database Management**:
  - Multi-database support (PostgreSQL, MySQL).
  - Configurable via `datasources.json` or Environment Variables.
  - Tools: `db_query`, `db_schema`, `db_inspect`.
- **Observability**:
  - **Logs**: Centralized log analysis (File + Elasticsearch).
  - **Metrics**: Prometheus metrics querying (`metrics.query`).
  - **Audit**: Comprehensive audit trail for all operations.
- **Testing & Diagnostics**:
  - **Load Testing**: `loadtest.run` (k6).
  - **Incident Analysis**: `incident.analyze` (Meta-tool combining logs, db, and metrics).

## 🛠️ Usage

### 1. Installation

```bash
git clone https://github.com/your-username/mcp-go.git
cd mcp-go
go mod download
make build
```

### 2. Configuration

Create a `.env` file (see `.env.example`):

```bash
# Server Metadata
SERVER_NAME="My Production MCP"
LOG_LEVEL=info

# Transport (stdio or http)
MCP_TRANSPORT=stdio
PORT=8080

# Elasticsearch (Optional)
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_API_KEY=your-api-key

# Prometheus (Optional)
PROMETHEUS_ADDR=http://localhost:9090
```

### 3. Running the Server

**Start with Docker (Recommended for Production)**
```bash
./deploy-to-docker.sh
# or manually:
docker-compose up -d
```
Services exposed:
- MCP Server: `http://localhost:8080`
- PostgreSQL: `localhost:5432`
- Elasticsearch: `http://localhost:9200`

**Mode: Stdio (Default - for Local Agents like Claude Desktop)**
```bash
./mcp-server
```

**Mode: HTTP (Server-Sent Events - for Remote Agents)**
```bash
MCP_TRANSPORT=http PORT=8080 ./mcp-server
```

### 4. DataSources Configuration

Create `datasources.json` in the project root to manage all data sources (databases, Elasticsearch, Prometheus, etc.):

```json
{
  "databases": [
    {
      "name": "default",
      "driver": "postgres",
      "dsn": "postgresql://user:pass@localhost:5432/main_db",
      "enabled": true
    },
    {
      "name": "analytics",
      "driver": "mysql",
      "dsn": "user:pass@tcp(analytics-db:3306)/events",
      "enabled": true
    }
  ],
  "elasticsearch": [
    {
      "name": "default",
      "url": "https://es.example.com",
      "api_key": "your-api-key",
      "enabled": true
    }
  ],
  "prometheus": [
    {
      "name": "default",
      "url": "http://prometheus:9090",
      "enabled": true
    }
  ]
}
```

See [DATASOURCES.md](DATASOURCES.md) for complete documentation.

---

## 🔌 Integration Guide

### Claude Desktop
Add to your config file:
```json
{
  "mcpServers": {
    "my-platform": {
      "command": "/absolute/path/to/mcp-server",
      "args": [],
      "env": {
        "MCP_TRANSPORT": "stdio"
      }
    }
  }
}
```

### OpenHands / Custom Agents (HTTP Mode)
1. Start server: `MCP_TRANSPORT=http ./mcp-server`
2. Connect using SSE Endpoint: `http://localhost:8080/sse`
3. Send JSON-RPC Requests to: `http://localhost:8080/messages`

### n8n Workflow Automation
**Full integration guide**: [N8N_INTEGRATION.md](N8N_INTEGRATION.md)

Quick example - HTTP Request node:
```json
POST {{ $json.url }}/message
Content-Type: application/json

{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "db_query",
    "arguments": {
      "database": "m3s-dev-mm",
      "query": "SELECT COUNT(*) FROM users"
    }
  }
}
```

See [N8N_INTEGRATION.md](N8N_INTEGRATION.md) for workflow templates, examples, and troubleshooting.

### Kiro / Other MCP Clients
If your client supports **Remote MCP** via SSE:
- **Server URL**: `http://localhost:8080/sse`
- **Port**: `8080`

If your client supports **Docker Execution**:
- **Command**: `docker run -i --rm --network host mcp-server`

---

## 📚 Available Tools

| Tool | Description |
|------|-------------|
| `db_query` | Execute SQL queries (SELECT, INSERT, UPDATE, etc.) |
| `db_schema` | View database table structures |
| `db_inspect` | Analyze query execution plans (EXPLAIN) |
| `logs_analyze` | Search logs in Files or Elasticsearch (supports multiple indices) |
| `metrics_query` | Query Prometheus metrics |
| `loadtest_run` | Run k6 load tests |
| `incident_analyze` | Diagnose system incidents (Meta-tool) |
| `platform_health` | Check MCP server status |

📖 **See [TOOLS_USAGE.md](TOOLS_USAGE.md) for detailed usage examples and patterns.**

## 🔒 Security

- **ReadOnly Mode**: Critical for production. Ensure DB users have restricted permissions.
- **API Keys**: Use `ELASTICSEARCH_API_KEY` instead of user/pass.
- **Audit Logging**: All tool executions are logged to `stderr`.

## 📜 License
MIT
