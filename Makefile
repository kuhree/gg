# Simple Makefile for a Go project

# Build the application
all: lint build test

build:
	@echo "Building..."
	@go build -o main cmd/gg/main.go


# Run the application
run:
	@go run cmd/gg/main.go


# Lint the application
lint:
	@echo "linting..."
	@gofmt -w -s -l ./..


# Test the application
test:
	@echo "Testing..."
	@go test ./... -v


# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main
	@rm -f tmp/**


# Live Reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run lint test clean watch docker-up docker-down
