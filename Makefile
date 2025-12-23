.PHONY: build test lint clean help setup

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o tsunami ./cmd/tsunami/

test: ## Run tests with coverage
	go test ./... -cover

lint: ## Run linter and fix issues
	mise exec -- golangci-lint run --fix

clean: ## Clean build artifacts
	rm -f tsunami coverage.out

setup: ## Install development tools
	mise install
