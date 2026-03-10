BINARY_NAME=mcp-server

build:
	@echo "Building optimized production binary..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) cmd/server/main.go

test:
	go test -v ./...

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)

docker-build:
	docker build -t mcp-platform-service:latest -f deployments/Dockerfile .

.PHONY: build test run clean docker-build
