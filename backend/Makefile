# Backend Testing Makefile
# Run from backend/ directory

.PHONY: test test-verbose test-coverage test-service test-integration test-websocket format lint clean

# Main test command
test:
	@echo "🧪 Running all backend tests..."
	go test -p=1 -parallel=1 ./test/...

test-backend:
	@echo "🧪 Running all backend tests..."
	go test -p=1 -parallel=1 ./test/...

# Test with verbose output
test-verbose:
	@echo "🧪 Running backend tests (verbose)..."
	go test -v -p=1 -parallel=1 ./test/...

# Test with coverage report
test-coverage:
	@echo "🧪 Running backend tests with coverage..."
	go test -v -p=1 -parallel=1 -coverprofile=coverage.out ./test/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report: coverage.html"

# Test specific categories
test-service:
	@echo "🔧 Running service layer tests..."
	go test ./test/service/...

test-integration:
	@echo "🔗 Running integration tests..."
	go test -p=1 -parallel=1 ./test/integration/...

test-websocket:
	@echo "🌐 Running WebSocket tests..."
	go test ./test/websocket/...

# Code quality
format:
	@echo "🎨 Formatting Go code..."
	gofmt -s -w .
	@echo "✅ Backend formatting complete"

lint:
	@echo "🔍 Running Go fmt..."
	go fmt ./...
	@echo "✅ Backend linting complete"

# Cleanup
clean:
	@echo "🧹 Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	go clean
	@echo "✅ Cleanup complete"

# Run specific test files (usage: make test-file FILE=test/service/game_service_test.go)
test-file:
	@echo "🎯 Running specific test file: $(FILE)"
	go test ./$(FILE)

# Run tests with race detection
test-race:
	@echo "🏁 Running tests with race detection..."
	go test -race -p=1 -parallel=1 ./test/...

# Watch for changes and run tests (requires entr: apt install entr)
test-watch:
	@echo "👀 Watching for Go file changes and running tests..."
	find . -name "*.go" | entr -c make test