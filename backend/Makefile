.PHONY: help build run test test-e2e test-e2e-auth test-e2e-data test-e2e-recyclebin clean restart-server

help:
	@echo "Available targets:"
	@echo "  build              - Build the server binary"
	@echo "  run                - Build and run the server"
	@echo "  restart-server     - Kill existing server and start fresh"
	@echo "  test               - Run all tests"
	@echo "  test-e2e           - Run all E2E tests"
	@echo "  test-e2e-auth      - Run authentication E2E tests"
	@echo "  test-e2e-data      - Run data API E2E tests"
	@echo "  test-e2e-recyclebin- Run recycle bin E2E tests"
	@echo "  clean              - Clean build artifacts"

build:
	@echo "Building server..."
	@go build -o bin/server cmd/server/main.go
	@echo "Build complete: bin/server"

run: build
	@echo "Starting server..."
	@set -a && source ../.env && set +a && PORT=3001 ./bin/server

restart-server:
	@echo "Restarting server..."
	@bash restart-server.sh

test:
	@echo "Running all tests..."
	@go test -v ./...

test-e2e:
	@echo "Running all E2E tests..."
	@cd ../tests/e2e && bash runner.sh

test-e2e-auth:
	@echo "Running auth E2E tests..."
	@go test -v ./e2e/auth

test-e2e-data:
	@echo "Running data E2E tests..."
	@go test -v ./e2e/data

test-e2e-recyclebin:
	@echo "Running recycle bin E2E tests..."
	@go test -v ./e2e/recyclebin

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean complete"
