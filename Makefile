# Simple Makefile for a Go project

# Load environment variables from .env
include .env
export $(shell sed 's/=.*//' .env)

GOOSE=goose
DB_DRIVER=postgres
MIGRATIONS_DIR=./migrations
DB_DSN=postgres://$(BLUEPRINT_DB_USERNAME):$(BLUEPRINT_DB_PASSWORD)@$(BLUEPRINT_DB_HOST):$(BLUEPRINT_DB_PORT)/$(BLUEPRINT_DB_DATABASE)?sslmode=disable&search_path=$(BLUEPRINT_DB_SCHEMA)

# Build the application
all: build test

build:
	@echo "Building..."
	
	
	@go build -o main cmd/api/main.go

# Run the application
run:
	@go run cmd/api/main.go
# Create DB container
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

# Shutdown DB container
docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

# Test the application
test:
	@echo "Testing..."
	@go test ./... -v
# Integrations Tests for the application
itest:
	@echo "Running integration tests..."
	@go test ./internal/database -v

# Clean the binary
clean:
	@echo "Cleaning..."
	@rm -f main

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

.PHONY: all build run test clean watch docker-run docker-down itest

# Commands

db-migrate: # make db-migrate
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" up

db-rollback:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" down

db-create: # make db-create name=create_users_table
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(name) sql

db-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" status

db-reset:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" reset

db-force: # make db-force version=20250417000123
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DB_DSN)" force $(version)