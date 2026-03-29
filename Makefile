# Makefile for API Backend Infinitrum

# Variables
APP_NAME=api-backend-infinitrum
GO_VERSION=1.21
DOCKER_IMAGE=$(APP_NAME):latest

# Default target
.DEFAULT_GOAL := help

# Help target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development targets
run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	go run main.go

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) main.go

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Testing targets
test: ## Run tests
	@echo "Running tests..."
	go test ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -cover ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	go test -v ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	go mod tidy

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

docker-compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo "Viewing docker-compose logs..."
	docker-compose logs -f

# Database targets
db-migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	migrate -path migrations -database "$(DATABASE_URL)" up

db-migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	migrate -path migrations -database "$(DATABASE_URL)" down

db-migrate-create: ## Create a new migration (usage: make db-migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)

# Code quality targets
fmt: ## Format Go code
	@echo "Formatting Go code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "Running golangci-lint..."
	golangci-lint run

# Development environment
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	cp .env.example .env
	@echo "Please edit .env file with your configuration"

dev-install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Production targets
prod-build: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o bin/$(APP_NAME) main.go

# Utility targets
logs: ## View application logs (if running with docker-compose)
	docker-compose logs -f api

restart: ## Restart the application (if running with docker-compose)
	docker-compose restart api

status: ## Check status of services
	docker-compose ps

# Git hooks
pre-commit: fmt vet test ## Run pre-commit checks

.PHONY: help run build clean test test-coverage test-verbose deps deps-update \
        docker-build docker-run docker-compose-up docker-compose-down docker-compose-logs \
        db-migrate-up db-migrate-down db-migrate-create \
        fmt vet lint dev-setup dev-install-tools prod-build \
        logs restart status pre-commit

