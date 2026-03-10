# DataSources Configuration

## Overview
`datasources.json` is the centralized configuration file for all data sources including:
- **Databases** (PostgreSQL, MySQL, SQLite)
- **Elasticsearch** (logs & search)
- **Prometheus** (metrics & monitoring)
- **Redis** (caching & sessions) - *optional*
- **S3** (object storage) - *optional*

## File Location
- **Development**: `./datasources.json`
- **Docker**: `/app/datasources.json` (mounted via docker-compose)

## Configuration Structure

```json
{
  "databases": [
    {
      "name": "default",
      "driver": "postgres",
      "dsn": "postgres://user:pass@host:5432/dbname?sslmode=disable",
      "enabled": true
    }
  ],
  "elasticsearch": [
    {
      "name": "default",
      "url": "https://your-elasticsearch-url.com",
      "api_key": "your-api-key-here",
      "enabled": true
    }
  ]
}
```

## Elasticsearch Configuration

### Using API Key (Recommended)
```json
"elasticsearch": [
  {
    "name": "default",
    "url": "https://elasticsearch.example.com",
    "api_key": "--",
    "enabled": true
  }
]
```

### Using Username/Password
```json
"elasticsearch": [
  {
    "name": "logs-cluster",
    "url": "https://elasticsearch.example.com",
    "user": "elastic",
    "pass": "password123",
    "enabled": true
  }
]
```

### Multiple Elasticsearch Clusters
```json
"elasticsearch": [
  {
    "name": "production",
    "url": "https://prod-es.example.com",
    "api_key": "prod-api-key",
    "enabled": true
  },
  {
    "name": "development",
    "url": "https://dev-es.example.com",
    "user": "elastic",
    "pass": "devpass",
    "enabled": false
  }
]
```
**Note**: The system will use the **first enabled** Elasticsearch cluster.

## Prometheus Configuration

### Basic Setup
```json
"prometheus": [
  {
    "name": "default",
    "url": "http://prometheus:9090",
    "enabled": true
  }
]
```

### Multiple Prometheus Instances
```json
"prometheus": [
  {
    "name": "production",
    "url": "https://prometheus-prod.example.com",
    "enabled": true
  },
  {
    "name": "development",
    "url": "http://localhost:9090",
    "enabled": false
  }
]
```

## Redis Configuration (Optional)

### Basic Setup
```json
"redis": [
  {
    "name": "cache",
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0,
    "enabled": true
  }
]
```

### Multiple Redis Instances
```json
"redis": [
  {
    "name": "cache",
    "host": "redis-cache",
    "port": 6379,
    "password": "cache-password",
    "db": 0,
    "enabled": true
  },
  {
    "name": "sessions",
    "host": "redis-sessions",
    "port": 6379,
    "password": "session-password",
    "db": 1,
    "enabled": true
  }
]
```

## S3 / Object Storage Configuration (Optional)

### AWS S3
```json
"s3": [
  {
    "name": "logs-storage",
    "endpoint": "s3.amazonaws.com",
    "region": "us-east-1",
    "access_key_id": "YOUR_ACCESS_KEY",
    "secret_access_key": "YOUR_SECRET_KEY",
    "bucket": "logs-bucket",
    "use_ssl": true,
    "enabled": true
  }
]
```

### MinIO (S3-compatible)
```json
"s3": [
  {
    "name": "minio-local",
    "endpoint": "minio:9000",
    "region": "us-east-1",
    "access_key_id": "minioadmin",
    "secret_access_key": "minioadmin",
    "bucket": "backups",
    "use_ssl": false,
    "enabled": true
  }
]
```

## Complete Example

```json
{
  "databases": [
    {
      "name": "default",
      "driver": "postgres",
      "dsn": "postgres://user:pass@host:5432/db?sslmode=disable",
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
  ],
  "redis": [
    {
      "name": "cache",
      "host": "redis",
      "port": 6379,
      "password": "",
      "db": 0,
      "enabled": false
    }
  ],
  "s3": [
    {
      "name": "backups",
      "endpoint": "minio:9000",
      "access_key_id": "minioadmin",
      "secret_access_key": "minioadmin",
      "bucket": "backups",
      "use_ssl": false,
      "enabled": false
    }
  ]
}
```

## Priority Order
1. **File-based config** (`datasources.json`) - highest priority
2. **Environment variables** - fallback if file not found or field empty

## Environment Variables (Fallback)
If you prefer environment variables or need to override:
```bash
```

## Docker Deployment
Set in docker-compose.yml:
```yaml
environment:
  - DATASOURCES_CONFIG=/app/datasources.json
volumes:
  - ./datasources.json:/app/datasources.json
```

## Backward Compatibility
The system still supports `DATABASES_CONFIG` environment variable and `databases.json` filename for backward compatibility.
