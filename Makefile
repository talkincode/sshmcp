.PHONY: help build test test-verbose test-coverage clean install run fmt vet lint deps

# Default target
.DEFAULT_GOAL := help

# Variable definitions
BINARY_NAME=sshx
BUILD_DIR=bin
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Go parameters
GOBASE=$(shell pwd)
GOBIN=$(GOBASE)/$(BUILD_DIR)
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

help: ## Show help information
	@echo "Available Make targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

build: ## Build binary
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(GOBIN)/$(BINARY_NAME) ./cmd/sshx
	@echo "Build complete: $(GOBIN)/$(BINARY_NAME)"

build-all: ## Build binaries for all platforms
	@echo "Building all platforms..."
	@mkdir -p $(BUILD_DIR)
	@echo "Building Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-linux-amd64 ./cmd/sshx
	@echo "Building Linux (arm64)..."
	GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-linux-arm64 ./cmd/sshx
	@echo "Building macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-amd64 ./cmd/sshx
	@echo "Building macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-darwin-arm64 ./cmd/sshx
	@echo "Building Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(GOBIN)/$(BINARY_NAME)-windows-amd64.exe ./cmd/sshx
	@echo "All platform builds complete!"

test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-short: ## Run unit tests (skip integration tests)
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./...

test-verbose: ## Run verbose tests
	@echo "Running verbose tests..."
	$(GOTEST) -v -race ./...

test-coverage: ## Run tests and generate coverage report
	@echo "Running tests and generating coverage..."
	$(GOTEST) -v -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./internal/app/...
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)
	@echo ""
	@echo "Generating HTML coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

test-app: ## Test app package only
	@echo "Testing app package..."
	$(GOTEST) -v ./internal/app/...

test-sshclient: ## Test sshclient package only
	@echo "Testing sshclient package..."
	$(GOTEST) -v ./internal/sshclient/...

clean: ## Clean build files and test cache
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "Clean complete!"

install: build ## Install to $GOPATH/bin and ~/bin
	@echo "Installing to system..."
	@if [ -n "$(GOPATH)" ] && [ -d "$(GOPATH)/bin" ]; then \
		cp $(GOBIN)/$(BINARY_NAME) $(GOPATH)/bin/; \
		echo "✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)"; \
	fi
	@if [ -d ~/bin ]; then \
		cp $(GOBIN)/$(BINARY_NAME) ~/bin/$(BINARY_NAME) && chmod +x ~/bin/$(BINARY_NAME); \
		echo "✓ Installed to ~/bin/$(BINARY_NAME)"; \
	fi
	@echo "Installation complete! You can now use '$(BINARY_NAME)' command"

uninstall: ## Uninstall from system
	@echo "Uninstalling..."
	@if [ -f "$(GOPATH)/bin/$(BINARY_NAME)" ]; then \
		rm -f $(GOPATH)/bin/$(BINARY_NAME); \
		echo "✓ Uninstalled from $(GOPATH)/bin"; \
	fi
	@if [ -f ~/bin/$(BINARY_NAME) ]; then \
		rm -f ~/bin/$(BINARY_NAME); \
		echo "✓ Uninstalled from ~/bin"; \
	fi
	@echo "Uninstall complete!"

run: build ## Build and run (show help)
	@echo "Running $(BINARY_NAME)..."
	@$(GOBIN)/$(BINARY_NAME) --help

fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) ./...
	@echo "Format complete!"

vet: ## Run go vet checks
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Check complete!"

lint: ## Run golangci-lint (requires installation)
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "Please install golangci-lint first: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
	@echo "Lint check complete!"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded!"

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	@echo "Dependencies tidied!"

vendor: ## Create vendor directory
	@echo "Creating vendor..."
	$(GOMOD) vendor
	@echo "Vendor created!"

check: fmt vet test ## Run all checks (format, vet, test)
	@echo "All checks passed!"

ci: deps check test-coverage ## CI/CD workflow (deps, check, coverage)
	@echo "CI workflow complete!"

tag:
	@echo "🏷️  Starting tag creation process..."
	@./scripts/tag.sh

dev: ## Development mode (install deps, format, test, build)
	@echo "Development mode..."
	@$(MAKE) deps
	@$(MAKE) fmt
	@$(MAKE) test
	@$(MAKE) build
	@echo "Development environment ready!"

release: clean test-coverage build-all ## Prepare release (clean, test, build all platforms)
	@echo "Preparing release..."
	@echo "All binaries located at: $(BUILD_DIR)/"
	@ls -lh $(BUILD_DIR)/
	@echo "Release ready!"

info: ## Show project information
	@echo "Project information:"
	@echo "  Name: $(BINARY_NAME)"
	@echo "  Go version: $(shell go version)"
	@echo "  Build directory: $(BUILD_DIR)"
	@echo "  Current path: $(GOBASE)"
	@echo ""
	@echo "Dependency statistics:"
	@go list -m all | wc -l | awk '{print "  Total dependencies: " $$1}'
	@echo ""
	@echo "Code statistics:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l | awk '{print "  Go files: " $$1}'
	@find . -name "*_test.go" -not -path "./vendor/*" | wc -l | awk '{print "  Test files: " $$1}'

watch: ## Watch file changes and auto-test (requires entr)
	@echo "Watching file changes..."
	@which entr > /dev/null || (echo "Please install entr first: brew install entr (macOS) or apt-get install entr (Linux)" && exit 1)
	@find . -name "*.go" -not -path "./vendor/*" | entr -c make test

.PHONY: all
all: clean deps fmt vet test build ## Complete build workflow
	@echo "Complete build done!"
