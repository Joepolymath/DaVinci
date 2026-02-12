# DaVinci Monorepo Makefile
# Centralized task runner for the polymath AI entity

# Directories
APPS_DIR := apps
LIBS_DIR := libs
INFRA_DIR := infra
TOOLS_DIR := tools

# Apps
SCRIBEQUERY_DIR := $(APPS_DIR)/scribequery
SYSTEM_AGENT_DIR := $(APPS_DIR)/system-agent
UI_DASHBOARD_DIR := $(APPS_DIR)/ui-dashboard
BRAIN_PROXY_DIR := $(APPS_DIR)/brain-proxy

# Libraries
PROTO_DIR := $(LIBS_DIR)/proto
SHARED_GO_DIR := $(LIBS_DIR)/shared-go
SHARED_TS_DIR := $(LIBS_DIR)/shared-ts

.PHONY: help
help: ## Show this help message
	@echo "DaVinci Monorepo Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# Proto & Contracts
# ============================================================================

.PHONY: proto-gen proto-clean
proto-gen: ## Generate code from Protobuf definitions
	@echo "Generating code from proto definitions..."
	@if [ -d "$(PROTO_DIR)" ]; then \
		cd $(PROTO_DIR) && find . -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative {} \; ; \
	else \
		echo "Proto directory not found. Creating structure..."; \
		mkdir -p $(PROTO_DIR); \
	fi

proto-clean: ## Clean generated proto code
	@find $(PROTO_DIR) -name "*.pb.go" -delete

# ============================================================================
# ScribeQuery (PDF RAG Engine - Go)
# ============================================================================

.PHONY: build-scribequery run-scribequery clean-scribequery test-scribequery
build-scribequery: ## Build ScribeQuery service
	@echo "Building ScribeQuery..."
	@go build -o bin/scribequery ./$(SCRIBEQUERY_DIR)/cmd/main.go

run-scribequery: ## Run ScribeQuery service
	@go run ./$(SCRIBEQUERY_DIR)/cmd/main.go

test-scribequery: ## Test ScribeQuery service
	@go test ./$(SCRIBEQUERY_DIR)/...

clean-scribequery: ## Clean ScribeQuery binaries
	@rm -rf $(SCRIBEQUERY_DIR)/bin

# ============================================================================
# System Agent (Telemetry & SysAdmin - Rust/Go)
# ============================================================================

.PHONY: build-system-agent run-system-agent clean-system-agent test-system-agent
build-system-agent: ## Build System Agent service
	@echo "Building System Agent..."
	@if [ -f "$(SYSTEM_AGENT_DIR)/Cargo.toml" ]; then \
		cd $(SYSTEM_AGENT_DIR) && cargo build --release; \
	elif [ -f "$(SYSTEM_AGENT_DIR)/go.mod" ]; then \
		cd $(SYSTEM_AGENT_DIR) && go build -o bin/system-agent ./cmd; \
	else \
		echo "System Agent not yet implemented"; \
	fi

run-system-agent: ## Run System Agent service
	@if [ -f "$(SYSTEM_AGENT_DIR)/Cargo.toml" ]; then \
		cd $(SYSTEM_AGENT_DIR) && cargo run; \
	elif [ -f "$(SYSTEM_AGENT_DIR)/go.mod" ]; then \
		cd $(SYSTEM_AGENT_DIR) && go run ./cmd; \
	else \
		echo "System Agent not yet implemented"; \
	fi

test-system-agent: ## Test System Agent service
	@if [ -f "$(SYSTEM_AGENT_DIR)/Cargo.toml" ]; then \
		cd $(SYSTEM_AGENT_DIR) && cargo test; \
	elif [ -f "$(SYSTEM_AGENT_DIR)/go.mod" ]; then \
		cd $(SYSTEM_AGENT_DIR) && go test ./...; \
	fi

clean-system-agent: ## Clean System Agent binaries
	@rm -rf $(SYSTEM_AGENT_DIR)/bin $(SYSTEM_AGENT_DIR)/target

# ============================================================================
# UI Dashboard (Unified Entity Interface - TS/Next.js)
# ============================================================================

.PHONY: build-ui run-ui dev-ui clean-ui install-ui
build-ui: ## Build UI Dashboard
	@echo "Building UI Dashboard..."
	@if [ -f "$(UI_DASHBOARD_DIR)/package.json" ]; then \
		cd $(UI_DASHBOARD_DIR) && npm run build; \
	else \
		echo "UI Dashboard not yet implemented"; \
	fi

run-ui: build-ui ## Run UI Dashboard (production)
	@cd $(UI_DASHBOARD_DIR) && npm start

dev-ui: ## Run UI Dashboard (development)
	@if [ -f "$(UI_DASHBOARD_DIR)/package.json" ]; then \
		cd $(UI_DASHBOARD_DIR) && npm run dev; \
	else \
		echo "UI Dashboard not yet implemented"; \
	fi

install-ui: ## Install UI Dashboard dependencies
	@if [ -f "$(UI_DASHBOARD_DIR)/package.json" ]; then \
		cd $(UI_DASHBOARD_DIR) && npm install; \
	fi

clean-ui: ## Clean UI Dashboard build artifacts
	@rm -rf $(UI_DASHBOARD_DIR)/.next $(UI_DASHBOARD_DIR)/node_modules/.cache

# ============================================================================
# Brain Proxy (LLM Gateway & Orchestrator - Go)
# ============================================================================

.PHONY: build-brain-proxy run-brain-proxy clean-brain-proxy test-brain-proxy
build-brain-proxy: ## Build Brain Proxy service
	@echo "Building Brain Proxy..."
	@if [ -f "$(BRAIN_PROXY_DIR)/go.mod" ]; then \
		cd $(BRAIN_PROXY_DIR) && go build -o bin/brain-proxy ./cmd; \
	else \
		echo "Brain Proxy not yet implemented"; \
	fi

run-brain-proxy: ## Run Brain Proxy service
	@if [ -f "$(BRAIN_PROXY_DIR)/go.mod" ]; then \
		cd $(BRAIN_PROXY_DIR) && go run ./cmd; \
	else \
		echo "Brain Proxy not yet implemented"; \
	fi

test-brain-proxy: ## Test Brain Proxy service
	@if [ -f "$(BRAIN_PROXY_DIR)/go.mod" ]; then \
		cd $(BRAIN_PROXY_DIR) && go test ./...; \
	fi

clean-brain-proxy: ## Clean Brain Proxy binaries
	@rm -rf $(BRAIN_PROXY_DIR)/bin

# ============================================================================
# Shared Libraries
# ============================================================================

.PHONY: build-libs test-libs clean-libs
build-libs: ## Build all shared libraries
	@echo "Building shared libraries..."
	@if [ -f "$(SHARED_GO_DIR)/go.mod" ]; then \
		cd $(SHARED_GO_DIR) && go build ./...; \
	fi

test-libs: ## Test all shared libraries
	@if [ -f "$(SHARED_GO_DIR)/go.mod" ]; then \
		cd $(SHARED_GO_DIR) && go test ./...; \
	fi

clean-libs: ## Clean shared library build artifacts
	@find $(LIBS_DIR) -name "*.pb.go" -delete

# ============================================================================
# Infrastructure
# ============================================================================

.PHONY: run-weaviate docker-up docker-down docker-build
run-weaviate: ## Run Weaviate vector database
	@docker run -d -p 8080:8080 -p 50051:50051 --name weaviate cr.weaviate.io/semitechnologies/weaviate:1.35.7

docker-up: ## Start all infrastructure services
	@if [ -f "$(INFRA_DIR)/docker-compose.yml" ]; then \
		cd $(INFRA_DIR) && docker-compose up -d; \
	else \
		echo "docker-compose.yml not found in infra/"; \
	fi

docker-down: ## Stop all infrastructure services
	@if [ -f "$(INFRA_DIR)/docker-compose.yml" ]; then \
		cd $(INFRA_DIR) && docker-compose down; \
	fi

docker-build: ## Build all Docker images
	@if [ -f "$(INFRA_DIR)/docker-compose.yml" ]; then \
		cd $(INFRA_DIR) && docker-compose build; \
	fi

# ============================================================================
# Development & Utilities
# ============================================================================

.PHONY: install-go-package install-rust-package install-ts-package
install-go-package: ## Install a Go package (usage: make install-go-package package=github.com/example/pkg)
	@if [ -z "$(package)" ]; then \
		echo "Usage: make install-go-package package=<package-path>"; \
		exit 1; \
	fi
	@echo "Installing $(package) in relevant Go modules..."
	@find $(APPS_DIR) $(LIBS_DIR) -name "go.mod" -exec sh -c 'cd $$(dirname {}) && go get $(package)' \;

install-rust-package: ## Install a Rust dependency (usage: make install-rust-package package=serde)
	@if [ -z "$(package)" ]; then \
		echo "Usage: make install-rust-package package=<crate-name>"; \
		exit 1; \
	fi
	@if [ -f "$(SYSTEM_AGENT_DIR)/Cargo.toml" ]; then \
		cd $(SYSTEM_AGENT_DIR) && cargo add $(package); \
	fi

install-ts-package: ## Install a TypeScript package (usage: make install-ts-package package=react)
	@if [ -z "$(package)" ]; then \
		echo "Usage: make install-ts-package package=<package-name>"; \
		exit 1; \
	fi
	@if [ -f "$(UI_DASHBOARD_DIR)/package.json" ]; then \
		cd $(UI_DASHBOARD_DIR) && npm install $(package); \
	fi

# ============================================================================
# Aggregate Commands
# ============================================================================

.PHONY: build-all run-all clean-all test-all
build-all: proto-gen build-libs build-scribequery build-system-agent build-brain-proxy build-ui ## Build all services and libraries

run-all: ## Run all services (in background)
	@echo "Starting all services..."
	@$(MAKE) run-scribequery &
	@$(MAKE) run-system-agent &
	@$(MAKE) run-brain-proxy &
	@$(MAKE) dev-ui

test-all: test-libs test-scribequery test-system-agent test-brain-proxy ## Run all tests

clean-all: clean-libs clean-scribequery clean-system-agent clean-brain-proxy clean-ui ## Clean all build artifacts
	@echo "Cleaned all build artifacts"