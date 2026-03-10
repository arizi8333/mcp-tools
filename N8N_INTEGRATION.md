# n8n Integration Guide

Panduan lengkap untuk menghubungkan n8n dengan MCP Server.

## 🔌 Koneksi n8n ke MCP Server

### Informasi Server MCP

- **Production URL**: `https://psops.pcsindonesia.com`
- **Endpoint**: `/message` atau `/sse`
- **Protocol**: JSON-RPC 2.0 over HTTP
- **Mode**: Stateless (tidak perlu sessionId)

---

## � Setup Manual di n8n (RECOMMENDED)

Cara paling mudah dan pasti work adalah setup manual:

### Step 1: Buat Workflow Baru

1. Login ke n8n
2. Klik **"+ Add workflow"**
3. Beri nama: "MCP Health Check Test"

### Step 2: Tambah Manual Trigger

1. Klik **"Add first step"**
2. Pilih **"On app event"** → **"Manual Trigger"** atau cari "Manual"
3. Node akan muncul dengan nama "When clicking 'Test workflow'"

### Step 3: Tambah HTTP Request Node

1. Klik **"+"** di sebelah kanan Manual Trigger node
2. Cari **"HTTP Request"**
3. Klik node "HTTP Request"

### Step 4: Configure HTTP Request

Di HTTP Request node, set parameter berikut:

**Basic Settings:**
- **Method**: `POST`
- **URL**: `https://psops.pcsindonesia.com/message`

**Authentication:**
- **Authentication**: `None` (pilih dari dropdown)

**Body:**
- Scroll ke bawah, cari section **"Body"**
- **Send Body**: Toggle ON (aktifkan)
- **Body Content Type**: Pilih `JSON`
- **Specify Body**: Pilih `Using JSON`
- **JSON**: Copy-paste kode ini:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "platform_health",
    "arguments": {}
  }
}
```

### Step 5: Test!

1. Klik tombol **"Test workflow"** di kanan atas
2. Tunggu beberapa detik
3. Harusnya muncul response hijau dengan data seperti:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Platform Status: Healthy\nUptime: ..."
      }
    ],
    "isError": false
  }
}
```

✅ **Jika sudah berhasil, berarti koneksi ke MCP server sudah OK!**

---

## 📊 Contoh Workflow Lainnya

### Example 1: Query Database

**HTTP Request Configuration:**
```
Method: POST
URL: https://psops.pcsindonesia.com/message
Body (JSON):
```

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "db_query",
    "arguments": {
      "database": "m3s-dev-mm",
      "query": "SELECT COUNT(*) as total_users FROM users LIMIT 1"
    }
  }
}
```

### Example 2: Search Logs dari Elasticsearch

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "logs_analyze",
    "arguments": {
      "index": "logstash-centralized-*",
      "pattern": "error",
      "time_range": "1h",
      "max_results": 50
    }
  }
}
```

### Example 3: List Semua Tools yang Tersedia

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/list",
  "params": {}
}
```

---

## 🔄 Workflow Template: Monitoring Alert

Buat workflow untuk monitoring otomatis:

```
[Schedule Trigger: Every 5 minutes]
    ↓
[HTTP Request: Check Platform Health]
    ↓
[IF Node: status !== "healthy"]
    ↓ (Yes - Ada masalah)
[Send Alert to Slack/Email]
```

**Schedule Trigger:**
- Type: Schedule Trigger
- Trigger Interval: Minutes
- Minutes Between Triggers: 5

**HTTP Request:**
- Same as Step 4 di atas

**IF Node:**
- Condition: `{{ $json.result.content[0].text }}` contains "Unhealthy"
- If true: Connect ke Slack/Email node

---

Gunakan **HTTP Request** node dengan konfigurasi:

```
Node: HTTP Request
Method: POST
URL: https://psops.pcsindonesia.com/message
```

**Headers:**
```
Content-Type: application/json
```

**Body (JSON):**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "roots": {
        "listChanged": true
      },
      "sampling": {}
    },
    "clientInfo": {
      "name": "n8n-workflow",
      "version": "1.0.0"
    }
  }
}
```

**Response yang diharapkan:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "protocolVersion": "2024-11-05",
    "capabilities": {
      "logging": {},
      "prompts": {
        "listChanged": true
      },
      "resources": {
        "subscribe": true,
        "listChanged": true
      },
      "tools": {
        "listChanged": true
      }
    },
    "serverInfo": {
      "name": "Production MCP Server",
      "version": "1.0.0"
    }
  }
}
```

---

### 2. List Available Tools

```
Node: HTTP Request
Method: POST
URL: https://psops.pcsindonesia.com/message
```

**Body:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/list",
  "params": {}
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "tools": [
      {
        "name": "platform_health",
        "description": "Check MCP server health",
        "inputSchema": {
          "type": "object",
          "properties": {}
        }
      },
      {
        "name": "db_query",
        "description": "Execute SQL queries",
        "inputSchema": {
          "type": "object",
          "properties": {
            "database": {"type": "string"},
            "query": {"type": "string"}
          },
          "required": ["database", "query"]
        }
      },
      {
        "name": "logs_analyze",
        "description": "Analyze logs from Elasticsearch or files",
        "inputSchema": {
          "type": "object",
          "properties": {
            "index": {"type": "string"},
            "pattern": {"type": "string"},
            "time_range": {"type": "string"}
          }
        }
      }
    ]
  }
}
```

---

### 3. Execute Tool (Call Database Query)

```
Node: HTTP Request
Method: POST
URL: https://psops.pcsindonesia.com/message
```

**Body:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "db_query",
    "arguments": {
      "database": "m3s-dev-mm",
      "query": "SELECT * FROM users LIMIT 5"
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Query results:\nid | name | email\n1 | John | john@example.com\n..."
      }
    ],
    "isError": false
  }
}
```

---

### 4. Analyze Logs dari Elasticsearch

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "logs_analyze",
    "arguments": {
      "index": "logstash-centralized-*",
      "pattern": "error",
      "time_range": "1h",
      "max_results": 50
    }
  }
}
```

---

## 🔄 n8n Workflow Templates

### Template 1: Database Query Automation

```
[Trigger: Webhook/Schedule]
    ↓
[HTTP Request: Initialize MCP]
    ↓
[HTTP Request: Query Database]
    ↓
[Process Results]
    ↓
[Send to Slack/Email]
```

### Template 2: Log Monitoring

```
[Schedule: Every 5 minutes]
    ↓
[HTTP Request: Initialize MCP]
    ↓
[HTTP Request: Analyze Logs]
    ↓
[IF Node: Has Errors?]
    ↓ (Yes)
[Create PagerDuty Alert]
```

### Template 3: Health Check Monitor

```
[Schedule: Every 1 minute]
    ↓
[HTTP Request: platform_health]
    ↓
[IF Node: Unhealthy?]
    ↓ (Yes)
[Send Alert via Telegram]
```

---

## 🛠️ Available Tools in MCP Server

| Tool Name | Description | Key Parameters |
|-----------|-------------|----------------|
| `db_query` | Execute SQL queries | `database`, `query` |
| `db_schema` | Get database schema | `database` |
| `db_inspect` | Analyze query plans | `database`, `query` |
| `logs_analyze` | Search logs (ES/File) | `index`, `pattern`, `time_range` |
| `metrics_query` | Query Prometheus metrics | `query`, `time` |
| `platform_health` | Server health check | - |
| `incident_analyze` | Diagnose incidents | `description` |
| `loadtest_run` | Run k6 load tests | `script`, `vus`, `duration` |

---

## 📊 Example n8n Workflow: Daily Database Report

### Node 1: Schedule Trigger
- **Type**: Schedule Trigger
- **Interval**: Every day at 8:00 AM

### Node 2: Initialize MCP
- **Type**: HTTP Request (POST)
- **URL**: `https://psops.pcsindonesia.com/message`
- **Body**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "clientInfo": {"name": "n8n-daily-report", "version": "1.0.0"}
  }
}
```

### Node 3: Query User Count
- **Type**: HTTP Request (POST)
- **URL**: `https://psops.pcsindonesia.com/message`
- **Body**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "db_query",
    "arguments": {
      "database": "m3s-dev-mm",
      "query": "SELECT COUNT(*) as total_users FROM users"
    }
  }
}
```

### Node 4: Query Today's Transactions
- **Type**: HTTP Request (POST)
- **Body**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "db_query",
    "arguments": {
      "database": "m3s-dev-ms",
      "query": "SELECT COUNT(*) as today_transactions FROM transactions WHERE DATE(created_at) = CURDATE()"
    }
  }
}
```

### Node 5: Format Report
- **Type**: Function
- **Code**:
```javascript
const userCount = $json["result"]["content"][0]["text"];
const txCount = $input.item(1).json["result"]["content"][0]["text"];

return {
  json: {
    report: `Daily Report:\nTotal Users: ${userCount}\nToday's Transactions: ${txCount}`,
    timestamp: new Date().toISOString()
  }
};
```

### Node 6: Send to Slack
- **Type**: Slack
- **Message**: `{{ $json.report }}`

---

## 🔍 Debugging Tips

### 1. Check Server Health
```bash
curl https://psops.pcsindonesia.com/health
```

### 2. Test dari Terminal
```bash
curl -X POST https://psops.pcsindonesia.com/message \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/list",
    "params": {}
  }'
```

### 3. Enable Debug di n8n
- Setting → Log Level → Debug
- Lihat execution logs untuk response detail

### 4. Common Errors

**Error: "Bad Request"**
- Pastikan `Content-Type: application/json`
- Pastikan JSON syntax benar (tidak ada trailing comma)

**Error: "Tool not found"**
- Gunakan `tools/list` untuk cek nama tool yang benar
- Tool names: `platform_health`, `db_query`, etc (gunakan underscore bukan dot)

**Error: "Database connection failed"**
- Cek di server logs: `docker logs mcp-server`
- Pastikan database configured di `datasources.json`

---

## 🚀 Quick Start: Minimal n8n Workflow

Copy JSON ini ke n8n (Import Workflow):

```json
{
  "nodes": [
    {
      "parameters": {
        "method": "POST",
        "url": "https://psops.pcsindonesia.com/message",
        "jsonParameters": true,
        "options": {},
        "bodyParametersJson": "{\n  \"jsonrpc\": \"2.0\",\n  \"id\": 1,\n  \"method\": \"tools/call\",\n  \"params\": {\n    \"name\": \"platform_health\",\n    \"arguments\": {}\n  }\n}"
      },
      "name": "MCP Health Check",
      "type": "n8n-nodes-base.httpRequest",
      "typeVersion": 1,
      "position": [250, 300]
    }
  ],
  "connections": {}
}
```

---

## 📝 Available Databases

Lihat [datasources.json](datasources.json) untuk database yang tersedia:

- **mdm** (PostgreSQL) - Master Data Management
- **m3s-dev-mm** (MySQL) - M3S Master MM
- **m3s-dev-ms** (MySQL) - M3S Master MS

---

## 🔗 Resources

- [MCP Protocol Spec](https://modelcontextprotocol.io/)
- [Tools Usage Guide](TOOLS_USAGE.md)
- [DataSources Configuration](DATASOURCES.md)
- [Server README](README.md)

---

## ⚡ Production Tips

1. **Use Credentials**: Store URL dan secrets di n8n Credentials
2. **Error Handling**: Tambahkan Error Trigger node untuk handle failures
3. **Rate Limiting**: Jangan spam requests, gunakan reasonable intervals
4. **Monitoring**: Setup alert untuk MCP server downtime
5. **Caching**: Cache `initialize` dan `tools/list` response jika memungkinkan

---

## 🆘 Troubleshooting

### n8n tidak bisa connect?

1. **Cek server running:**
```bash
curl https://psops.pcsindonesia.com/health
# Should return: OK
```

2. **Cek CORS:**
Server sudah enable CORS untuk all origins (`Access-Control-Allow-Origin: *`)

3. **Cek dari Docker:**
Jika n8n di Docker, pastikan bisa akses external URLs

4. **Test dengan curl dulu:**
Sebelum setup di n8n, test dengan curl untuk verify endpoint works

### Response kosong?

- Cek `id` di request dan response harus match
- Pastikan method name benar: `tools/call` bukan `tool/call`
- Cek server logs: `journalctl -u mcp-server -f` atau `docker logs mcp-server`

---

**Need help?** Check server logs atau contact IT team.
