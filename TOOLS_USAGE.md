# MCP Tools Usage Guide

## `logs_analyze` - Log Analysis Tool

Search for patterns, errors, or summaries in log files or Elasticsearch.

### Elasticsearch Index Patterns

Elasticsearch supports multiple ways to specify indices:

#### 1. Single Index
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logstash-centralized-development",
    "pattern": "error",
    "time_range": "1h"
  }
}
```

#### 2. Wildcard Pattern (Recommended)
Search across all indices matching a pattern:
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logstash-*",
    "pattern": "exception",
    "time_range": "24h",
    "max_results": 100
  }
}
```

Common wildcard patterns:
- `logs-*` - All indices starting with "logs-"
- `*-2024` - All indices ending with "-2024"
- `logstash-*-development` - All development logstash indices

#### 3. Multiple Indices (Comma-Separated)
Search specific multiple indices:
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logs-app,logs-api,logs-worker",
    "pattern": "timeout",
    "time_range": "12h"
  }
}
```

#### 4. Mixed Pattern (Wildcards + Specific)
Combine specific indices with patterns:
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logstash-centralized-*,audit-logs-2024",
    "pattern": "failed login",
    "time_range": "7d"
  }
}
```

### Time Range Options

- `1h` - Last 1 hour
- `6h` - Last 6 hours
- `12h` - Last 12 hours
- `24h` - Last 24 hours (default)
- `7d` - Last 7 days
- `30d` - Last 30 days

### Pattern Search

Supports wildcards and boolean operators:
- `error` - Simple keyword
- `*exception*` - Wildcard pattern
- `error OR warning` - Boolean OR
- `"exact phrase"` - Exact match
- `status:500` - Field-specific search

### Complete Examples

#### Example 1: Search all production logs for errors
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logs-*-production",
    "pattern": "error OR exception OR fatal",
    "time_range": "1h",
    "max_results": 50
  }
}
```

#### Example 2: Search specific applications
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "app-frontend,app-backend,app-api",
    "pattern": "status:500",
    "time_range": "24h",
    "max_results": 100
  }
}
```

#### Example 3: Search all centralized logs
```json
{
  "name": "logs_analyze",
  "arguments": {
    "index": "logstash-centralized-*",
    "pattern": "database connection",
    "time_range": "6h"
  }
}
```

#### Example 4: File-based log search
```json
{
  "name": "logs_analyze",
  "arguments": {
    "path": "/var/log/application.log",
    "pattern": "ERROR",
    "max_results": 20
  }
}
```

---

## `db_query` - Database Query Tool

Execute SQL queries against configured databases.

### Basic Query
```json
{
  "name": "db_query",
  "arguments": {
    "database": "default",
    "query": "SELECT * FROM users LIMIT 10"
  }
}
```

### Multiple Databases
If you have multiple databases configured in `datasources.json`:
```json
{
  "name": "db_query",
  "arguments": {
    "database": "analytics",
    "query": "SELECT COUNT(*) FROM events WHERE created_at > NOW() - INTERVAL 1 DAY"
  }
}
```

---

## `db_schema` - Database Schema Inspector

Get schema information for a database.

```json
{
  "name": "db_schema",
  "arguments": {
    "database": "default"
  }
}
```

---

## `db_inspect` - Database Query Analyzer

Analyze SQL queries with EXPLAIN.

```json
{
  "name": "db_inspect",
  "arguments": {
    "database": "default",
    "query": "SELECT * FROM orders WHERE customer_id = 123"
  }
}
```

---

## `metrics_query` - Prometheus Metrics Query

Execute PromQL queries.

```json
{
  "name": "metrics_query",
  "arguments": {
    "query": "rate(http_requests_total[5m])",
    "time": "now"
  }
}
```

---

## `platform_health` - Platform Health Check

Get server health and uptime.

```json
{
  "name": "platform_health",
  "arguments": {}
}
```

---

## Best Practices

### 1. Start Narrow, Then Expand
Start with specific index and expand if needed:
```
logstash-centralized-development → logstash-* → *
```

### 2. Use Time Ranges Wisely
- Recent issues: `1h` or `6h`
- Pattern analysis: `24h` or `7d`
- Long-term trends: `30d`

### 3. Limit Results
Always set `max_results` to avoid overwhelming responses:
- Quick check: 10-20 results
- Analysis: 50-100 results
- Deep dive: 200-500 results

### 4. Combine with Other Tools
```
1. logs_analyze → Find errors
2. db_query → Check affected data
3. metrics_query → Verify impact
```
