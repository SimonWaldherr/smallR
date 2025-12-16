GOROOT := $(shell go env GOROOT)
OUT := demo/smallr.wasm
PKG := ./cmd/smallr-wasm

.PHONY: all build-wasm copy-wasm-exec clean serve test test-verbose test-coverage

all: build-wasm copy-wasm-exec

# Build WebAssembly module named smallr.wasm into the demo directory
build-wasm:
	@echo "Building WASM -> $(OUT)"
	GOOS=js GOARCH=wasm go build -o $(OUT) $(PKG)

# Copy Go's wasm_exec.js into the demo folder (if present)
copy-wasm-exec:
	@if [ -f "$(GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$(GOROOT)/misc/wasm/wasm_exec.js" demo/wasm_exec.js && echo "Copied wasm_exec.js"; \
	else \
		echo "wasm_exec.js not found in $(GOROOT)/misc/wasm — skipping copy"; \
	fi

clean:
	@echo "Removing $(OUT)"
	@rm -f $(OUT)

# Serve the demo directory on :8000 (Python 3)
serve:
	@echo "Serving demo at http://localhost:8000"
	@cd demo && python3 -m http.server 8000

# Run all tests
test:
	@echo "Running tests..."
	@go test ./internal/... -v

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	@go test ./internal/... -v -count=1

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test ./internal/... -cover -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run specific package tests
test-lexer:
	@echo "Testing lexer..."
	@go test ./internal/lexer -v

test-parser:
	@echo "Testing parser..."
	@go test ./internal/parser -v

test-rt:
	@echo "Testing runtime..."
	@go test ./internal/rt -v
