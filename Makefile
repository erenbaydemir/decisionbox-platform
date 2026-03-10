.PHONY: help infra-up infra-down build test lint clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Infrastructure ---

infra-up: ## Start infrastructure (MongoDB)
	docker compose up -d

infra-down: ## Stop infrastructure
	docker compose down

infra-reset: ## Stop infrastructure and remove volumes
	docker compose down -v

# --- Build ---

build: build-agent build-api ## Build all services

build-agent: ## Build the DecisionBox Agent
	cd services/agent && go build -o ../../bin/decisionbox-agent .

build-api: ## Build the DecisionBox API
	cd services/api && go build -o ../../bin/decisionbox-api .

# --- Test ---

test: ## Run all tests
	cd libs/go-common && go test ./...
	cd providers/llm/claude && go test ./...
	cd providers/warehouse/bigquery && go test ./...
	cd domain-packs/gaming/go && go test ./...

test-coverage: ## Run tests with coverage
	cd libs/go-common && go test -cover ./...

# --- Lint ---

lint: ## Run linters
	@echo "TODO: add golangci-lint"

# --- Clean ---

clean: ## Remove build artifacts
	rm -rf bin/
