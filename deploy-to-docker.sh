#!/bin/bash
set -e

# Build the Docker image
echo "Building Docker image..."
docker build -t mcp-server:latest .

# Check if docker compose (v2) is installed
if docker compose version &> /dev/null; then
    echo "Starting services with docker compose..."
    docker compose up -d
    echo "Services started! MCP Server running on http://localhost:8080"
elif command -v docker-compose &> /dev/null; then
    echo "Starting services with docker-compose..."
    docker-compose up -d
    echo "Services started! MCP Server running on http://localhost:8080"
else
    echo "docker-compose not found. Please install docker-compose or run manually:"
    echo "docker run -p 8080:8080 mcp-server:latest"
fi
