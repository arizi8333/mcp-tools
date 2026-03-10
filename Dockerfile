# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make build-base

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o mcp-server cmd/server/main.go

# Production Stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies (ca-certificates for HTTPS)
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/mcp-server .

# Create directory for config files
RUN mkdir -p /etc/mcp

# Expose ports (for HTTP mode)
EXPOSE 8080 8180

# Default command (can be overridden by docker run arguments)
ENTRYPOINT ["./mcp-server"]
