# Makefile for wistia-downloader

.PHONY: build clean test run help install-deps

APP_NAME := wistia-downloader
VERSION := $(shell date +%Y%m%d-%H%M%S)
BUILD_DIR := build
SRC_DIR := src

# Default target
help:
	@echo "Available targets:"
	@echo "  build         - Build binaries for all platforms"
	@echo "  build-local   - Build binary for current platform only"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  run           - Run the application locally"
	@echo "  install-deps  - Install Go dependencies"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION       - Set version (default: timestamp)"

# Build for all platforms using the build script
build:
	@echo "Building $(APP_NAME) for all platforms..."
	./build.sh

# Build for current platform only
build-local:
	@echo "Building $(APP_NAME) for current platform..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./$(SRC_DIR)
	@echo "Built $(BUILD_DIR)/$(APP_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Run tests
test:
	go test ./...

# Run the application
run:
	go run ./$(SRC_DIR) $(ARGS)

# Install dependencies
install-deps:
	go mod tidy
	go mod download

# Development build (with debug info)
build-dev:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME)-dev ./$(SRC_DIR)
	@echo "Built development binary: $(BUILD_DIR)/$(APP_NAME)-dev"
