.PHONY: all build run test clean frontend backend docker dev

# Variables
BINARY_NAME=stale
VERSION?=0.1.0
BUILD_DIR=./build
FRONTEND_DIR=./ui

all: build

# Frontend build
frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

# Backend build (requires frontend to be built first)
backend:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/stale

# Full build
build: frontend backend

# Run in development mode
run:
	go run ./cmd/stale

# Run frontend dev server
dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

# Run tests
test:
	go test -v -race ./...

test-frontend:
	cd $(FRONTEND_DIR) && npm test

# Docker build
docker:
	docker build -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	docker run -d -p 8080:8080 -v stale-data:/data --name stale $(BINARY_NAME):$(VERSION)

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules

# Install dependencies
setup:
	go mod download
	cd $(FRONTEND_DIR) && npm ci

# Lint
lint:
	golangci-lint run
	cd $(FRONTEND_DIR) && npm run lint
