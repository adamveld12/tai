# TAI (Terminal AI) Makefile

# Variables
BINARY_NAME=tai
MAIN_PACKAGE=./cmd/tai
BUILD_DIR=./build
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

# Build for multiple platforms
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Install the application
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PACKAGE)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -cover ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Watch and run tests on file changes
.PHONY: check-watch
check-watch:
	@echo "Watching for changes and running tests..."
	@if command -v air >/dev/null 2>&1; then \
		air -c .air.test.toml; \
	else \
		echo "Air not found. Install it with: go install github.com/cosmtrek/air@latest"; \
		exit 1; \
	fi

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		@echo "golangci-lint not found. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		exit 1; \
	fi

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet the code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all quality checks
.PHONY: fast-check
fast-check: vet lint test test-race

# Run all quality checks
.PHONY: check
check: fmt vet lint test test-race

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Run the application
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PACKAGE)

# run lmstudio
.PHONY: lmstudio
lmstudio:
	lmstudio server start

# Development server (with live reload if available)
.PHONY: dev
dev:
	go run $(MAIN_PACKAGE)

# Development server (with live reload if available)
.PHONY: dev-reload
dev-reload:
	@if command -v air >/dev/null 2>&1; then \
    	air -build.cmd 'make build' -build.bin ./build/tai; \
	else \
		echo "Air not found. Running without live reload..."; \
		go run $(MAIN_PACKAGE); \
	fi

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build for multiple platforms"
	@echo "  install       - Install the application"
	@echo "  test          - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report and HTML output"
	@echo "  check-watch   - Watch files and run tests on changes (requires air)"
	@echo "  bench         - Run benchmarks"
	@echo "  lint          - Run linter (requires golangci-lint)"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  fast-check    - Run quality checks without formatting"
	@echo "  check         - Run all quality checks (fmt, vet, lint, test, test-race)"
	@echo "  clean         - Clean build artifacts and coverage files"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  update-deps   - Update all dependencies"
	@echo "  run           - Run the application"
	@echo "  lmstudio      - Start LMStudio server"
	@echo "  dev           - Run the application (same as run)"
	@echo "  dev-reload    - Run with live reload (requires air)"
	@echo "  help          - Show this help message"
