# Simple Makefile for a Go project

# Build the application
all: lint build test
templ-install:
	@if ! command -v templ > /dev/null; then \
		read -p "Go's 'templ' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
		if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
			go install github.com/a-h/templ/cmd/templ@latest; \
			if [ ! -x "$$(command -v templ)" ]; then \
				echo "templ installation failed. Exiting..."; \
				exit 1; \
			fi; \
		else \
			echo "You chose not to install templ. Exiting..."; \
			exit 1; \
		fi; \
	fi

build: templ-install
	@echo "Building..."
	@templ generate
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

.PHONY: all build run lint test clean watch templ-install docker-up docker-down
