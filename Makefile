# NexusCRM Makefile
# Provides common build and development commands

.PHONY: generate build verify-generated lint test clean

# Default target
all: build

# Generate all code from system_tables.json (single source of truth)
generate:
	@echo "🔄 Generating code from system_tables.json..."
	cd backend && go run ./cmd/codegen
	@echo "✅ Generation complete"

# Build all components
build: generate
	@echo "🔨 Building backend..."
	cd backend && go build ./...
	@echo "🔨 Building MCP..."
	cd mcp && go build ./...
	@echo "🔨 Type-checking frontend..."
	npm run lint
	@echo "✅ Build complete"

# Verify generated code is up-to-date (used in CI)
verify-generated:
	@echo "🔍 Verifying generated code is up-to-date..."
	cd backend && go run ./cmd/codegen
	@git diff --exit-code backend/pkg/constants/z_generated_tables.go || (echo "❌ z_generated_tables.go is out of sync" && exit 1)
	@git diff --exit-code backend/pkg/constants/z_generated_fields.go || (echo "❌ z_generated_fields.go is out of sync" && exit 1)
	@git diff --exit-code backend/internal/domain/models/z_generated.go || (echo "❌ models/z_generated.go is out of sync" && exit 1)
	@git diff --exit-code frontend/src/generated-schema.ts || (echo "❌ generated-schema.ts is out of sync" && exit 1)
	@git diff --exit-code mcp/pkg/models/z_generated.go || (echo "❌ mcp/z_generated.go is out of sync" && exit 1)
	@echo "✅ Generated files are up-to-date"

# Run linting
lint:
	@echo "🔍 Linting..."
	cd backend && go vet ./...
	npm run lint
	@echo "✅ Lint complete"

# Run all tests
test:
	@echo "🧪 Running backend tests..."
	cd backend && go test ./...
	@echo "🧪 Running MCP tests..."
	cd mcp && go test ./...
	@echo "✅ Tests complete"

# Clean generated files (for fresh regeneration)
clean:
	@echo "🧹 Cleaning generated files..."
	rm -f backend/pkg/constants/z_generated_tables.go
	rm -f backend/pkg/constants/z_generated_fields.go
	rm -f backend/internal/domain/models/z_generated.go
	rm -f frontend/src/generated-schema.ts
	rm -f mcp/pkg/models/z_generated.go
	@echo "✅ Clean complete"

# Help
help:
	@echo "NexusCRM Makefile Commands:"
	@echo "  make generate          - Generate code from system_tables.json"
	@echo "  make build            - Generate and build all components"
	@echo "  make verify-generated - Verify generated files are up-to-date (for CI)"
	@echo "  make lint             - Run linting"
	@echo "  make test             - Run all tests"
	@echo "  make clean            - Remove generated files"
