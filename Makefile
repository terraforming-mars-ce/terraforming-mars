# Terraforming Mars - Unified Development Makefile
# Run from project root directory

.PHONY: help run frontend backend kill lint typecheck test test-backend test-frontend test-verbose test-coverage clean build format format-backend format-frontend install-cli generate prepare-for-commit deploy-pi mcp-setup bot-setup bot-run deps

# Default target - show help
help:
	@echo "🚀 Terraforming Mars Development Commands"
	@echo ""
	@echo "🎯 Main Commands:"
	@echo "  make run          - Run both frontend and backend with hot reload"
	@echo "  make frontend     - Run frontend development server (port 3000)"
	@echo "  make backend      - Build and run backend (port 3001)"
	@echo "  make kill         - Kill all frontend and backend development processes"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test         - Run all tests (backend + frontend)"
	@echo "  make test-backend - Run backend tests only"
	@echo "  make test-verbose - Run backend tests with verbose output"
	@echo "  make test-coverage- Run backend tests with coverage report"
	@echo ""
	@echo "🔧 Code Quality:"
	@echo "  make lint              - Run all linters (backend + frontend)"
	@echo "  make typecheck         - Run TypeScript type checking"
	@echo "  make format            - Format all code (Go + TypeScript)"
	@echo "  make generate          - Generate TypeScript types from Go structs"
	@echo "  make prepare-for-commit- Format and lint before committing"
	@echo ""
	@echo "🏗️  Build & Deploy:"
	@echo "  make build        - Build production binaries"
	@echo "  make clean        - Clean build artifacts"
	@echo ""

# Main development commands
run:
	@echo "🚀 Starting Terraforming Mars (frontend + backend with hot reload)..."
	@echo "   Frontend: http://localhost:3000"
	@echo "   Backend:  http://localhost:3001"
	@echo "   Press Ctrl+C to stop both servers"
	@echo ""
	@trap 'kill 0' SIGINT; \
		(cd backend && TM_REPO_PATH=../ $(shell go env GOPATH)/bin/air) & \
		(cd frontend && npm start) & \
		wait

frontend:
	@echo "🎨 Starting frontend development server..."
	cd frontend && npm start

backend: build-backend
	@echo "🚀 Running backend (port 3001)..."
	cd backend && TM_REPO_PATH=../ ./bin/server

kill:
	@echo "🛑 Killing all development servers..."
	./kill-servers.sh

# Testing commands
test: test-backend

test-backend:
	@echo "🧪 Running backend tests..."
	cd backend && go test ./test/...

test-frontend:
	@echo "🧪 Running frontend tests..."
	@echo "⚠️  No test script found in frontend package.json"
	@echo "ℹ️  Running linter instead..."
	cd frontend && npm run lint

test-verbose:
	@echo "🧪 Running backend tests (verbose)..."
	cd backend && go test -v ./test/...

test-coverage:
	@echo "🧪 Running backend tests with coverage..."
	cd backend && go test -v -coverprofile=coverage.out -coverpkg=./internal/... ./test/...
	@cd backend && if [ -s coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html && \
		echo "📊 Coverage report generated: backend/coverage.html"; \
	else \
		echo "⚠️ No coverage data generated - skipping HTML report"; \
	fi
	@echo "✅ Test coverage completed"

# Quick test commands for development
test-quick:
	@echo "⚡ Running quick test suite..."
	@cd backend && go test ./test/service/... && echo "✅ Service tests passed" || echo "❌ Service tests failed"
	@cd backend && go test ./test/delivery/websocket/hub_test.go && echo "✅ Hub tests passed" || echo "❌ Hub tests failed"
	@cd backend && go test ./test/delivery/websocket/message_test.go && echo "✅ Message tests passed" || echo "❌ Message tests failed"
	@cd backend && go test ./test/delivery/websocket/client_test.go && echo "✅ Client tests passed" || echo "❌ Client tests failed"

# Code quality commands
lint: lint-backend lint-frontend typecheck

typecheck:
	@echo "🔍 Running TypeScript type checking..."
	cd frontend && npm run typecheck
	@echo "✅ Type checking complete"

lint-backend:
	@echo "🔍 Running backend linting (Go fmt)..."
	cd backend && go fmt ./...
	@echo "✅ Backend formatting complete"

lint-frontend:
	@echo "🔍 Running frontend linting (oxlint)..."
	cd frontend && npm run lint
	@echo "✅ Frontend linting complete"

format: format-backend format-frontend

format-backend:
	@echo "🎨 Formatting backend Go code..."
	cd backend && find . -name "*.go" -exec gofmt -s -w {} \;
	@echo "✅ Backend formatting complete"

format-frontend:
	@echo "🎨 Formatting frontend TypeScript code..."
	cd frontend && npm run format:write
	@echo "✅ Frontend formatting complete"

# Pre-commit preparation
prepare-for-commit: format lint
	@echo "✅ Ready to commit"

# Build and deployment
build: build-backend build-frontend

build-backend:
	@echo "🏗️  Building backend binary..."
	cd backend && go build -o bin/server cmd/server/main.go
	@echo "✅ Backend binary: backend/bin/server"

build-frontend:
	@echo "🏗️  Building frontend for production..."
	cd frontend && npm run build
	@echo "✅ Frontend build: frontend/dist/"

# Cleanup
clean:
	@echo "🧹 Cleaning build artifacts..."
	cd backend && rm -f bin/server bin/tm coverage.out coverage.html
	cd frontend && rm -rf dist build
	cd backend && go clean
	@echo "✅ Cleanup complete"

# Install dependencies
deps:
	cd backend && go mod tidy
	cd frontend && npm install

# Development helpers
dev-setup:
	@echo "🔧 Setting up development environment..."
	go install github.com/air-verse/air@latest
	cd backend && go mod tidy
	cd frontend && npm install
	@echo "✅ Development setup complete"

# Type generation
generate:
	@echo "🔄 Generating TypeScript types from Go structs..."
	cd backend && tygo generate
	@echo "✅ TypeScript types generated"

# MCP server setup
mcp-setup:
	@echo "Setting up MCP server..."
	cd mcp-server && npm install
	@echo "MCP server ready. Restart Claude Code to pick up .mcp.json"

# Raspberry Pi deployment
deploy-pi:
	./scripts/deploy-pi.sh

# Claude Bot
bot-setup:
	@echo "Setting up claude-bot..."
	cd claude-bot && go mod tidy
	@echo "Claude bot ready"

bot-run:
	@if [ -z "$(GAME)" ]; then \
		echo "Usage: make bot-run GAME=<game-id> [NAME='Claude Bot'] [MODEL=sonnet]"; \
		exit 1; \
	fi
	cd claude-bot && go run cmd/bot/main.go \
		--game "$(GAME)" \
		--name "$(or $(NAME),Claude Bot)" \
		--model "$(or $(MODEL),sonnet)"

# Watch for changes (requires entr: apt install entr)
test-watch:
	@echo "👀 Watching for Go file changes and running tests..."
	cd backend && find . -name "*.go" | entr -c make test-quick