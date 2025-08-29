# Makefile for Repo Onboarding Copilot

# Variables
APP_NAME := repo-onboarding-copilot
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE) -X main.CommitHash=$(COMMIT_HASH)"

# Build directories
BUILD_DIR := build
DIST_DIR := dist

# Go build flags
GO_BUILD_FLAGS := -trimpath $(LDFLAGS)

# Default target
.DEFAULT_GOAL := build

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Clean build artifacts
.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	go clean -cache

# Install dependencies
.PHONY: deps
deps: ## Install dependencies
	go mod download
	go mod tidy

# Build for current platform
.PHONY: build
build: deps ## Build for current platform
	mkdir -p $(BUILD_DIR)
	go build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(APP_NAME) ./cmd

# Cross-platform builds
.PHONY: build-all
build-all: build-darwin build-linux build-windows ## Build for all platforms

.PHONY: build-darwin
build-darwin: deps ## Build for macOS (Darwin)
	mkdir -p $(DIST_DIR)/darwin-amd64 $(DIST_DIR)/darwin-arm64
	GOOS=darwin GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/darwin-amd64/$(APP_NAME) ./cmd
	GOOS=darwin GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/darwin-arm64/$(APP_NAME) ./cmd

.PHONY: build-linux
build-linux: deps ## Build for Linux
	mkdir -p $(DIST_DIR)/linux-amd64 $(DIST_DIR)/linux-arm64
	GOOS=linux GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/linux-amd64/$(APP_NAME) ./cmd
	GOOS=linux GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/linux-arm64/$(APP_NAME) ./cmd

.PHONY: build-windows
build-windows: deps ## Build for Windows
	mkdir -p $(DIST_DIR)/windows-amd64 $(DIST_DIR)/windows-arm64
	GOOS=windows GOARCH=amd64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/windows-amd64/$(APP_NAME).exe ./cmd
	GOOS=windows GOARCH=arm64 go build $(GO_BUILD_FLAGS) -o $(DIST_DIR)/windows-arm64/$(APP_NAME).exe ./cmd

# Test targets
.PHONY: test
test: ## Run tests
	go test -v -race -cover ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Quality checks
.PHONY: lint
lint: ## Run linter
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@export PATH="$(shell go env GOPATH)/bin:$$PATH"; golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	go mod tidy

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# Security checks
.PHONY: security
security: ## Run security checks
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	@export PATH="$(shell go env GOPATH)/bin:$$PATH"; gosec ./...

# Full quality check
.PHONY: check
check: fmt vet lint test security ## Run all quality checks

# Development
.PHONY: run
run: build ## Build and run the application
	./$(BUILD_DIR)/$(APP_NAME) --help

.PHONY: dev
dev: ## Run in development mode
	go run ./cmd --help

# Install
.PHONY: install
install: ## Install to GOPATH/bin
	go install $(GO_BUILD_FLAGS) ./cmd

# Version info
.PHONY: version
version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Commit: $(COMMIT_HASH)"