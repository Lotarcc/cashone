# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=cashone
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./cmd
MIGRATE_PATH=./cmd/migrate

# Version information
VERSION ?= $(shell git describe --tags --always --dirty || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

.PHONY: all init setup build run clean build-linux docker-build docker-run docker-down db-migrate db-rollback db-status db-seed db-test db-new

# Main targets
all: test build

# Initialization and setup
init:
	./scripts/init-env.sh

setup: init
	./scripts/setup.sh

# Version management
version:
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

update-version:
	./scripts/build-version.sh

# Development targets
dev: setup db-up
	$(MAKE) run

dev-live:
	./scripts/dev.sh

dev-reset: clean db-reset
	$(MAKE) dev-live

dev-install:
	go install github.com/air-verse/air@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golint/cmd/golint@latest

# Documentation targets
docs:
	./scripts/generate-docs.sh

serve-docs: docs
	@echo "Opening Swagger UI..."
	@which xdg-open > /dev/null && xdg-open http://localhost:8081/swagger/index.html || \
	which open > /dev/null && open http://localhost:8081/swagger/index.html || \
	echo "Please open http://localhost:8081/swagger/index.html in your browser"

# Build targets
build: update-version
	cd $(MAIN_PATH) && $(GOBUILD) -o ../../bin/$(BINARY_NAME) -v
	cd $(MIGRATE_PATH) && $(GOBUILD) -o ../../../bin/migrate -v

build-linux: update-version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v
	cd $(MIGRATE_PATH) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o ../../../bin/migrate -v

# Run targets
run:
	cd $(MAIN_PATH) && $(GORUN) main.go

lint:
	golangci-lint run

check: lint test-coverage

# Clean targets
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)
	rm -f bin/migrate
	rm -rf coverage/
	rm -rf docs/
	rm -rf tmp/

# Docker targets
docker-build: update-version
	docker build -t $(BINARY_NAME) .

docker-run:
	docker compose up

docker-down:
	docker compose down

# Database targets
db-up:
	docker compose up -d db
	@echo "Waiting for database to be ready..."
	@sleep 5
	$(MAKE) db-migrate

db-down:
	docker compose stop db

db-migrate:
	cd $(MIGRATE_PATH) && $(GORUN) main.go -command up

db-rollback:
	cd $(MIGRATE_PATH) && $(GORUN) main.go -command down

db-status:
	cd $(MIGRATE_PATH) && $(GORUN) main.go -command status

db-new:
	@if [ "$(name)" = "" ]; then \
		echo "Error: name is required. Use: make db-new name=description_of_migration"; \
		exit 1; \
	fi
	@timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$$(echo $(name) | tr ' ' '_').sql"; \
	down_filename="migrations/$${timestamp}_$$(echo $(name) | tr ' ' '_')_down.sql"; \
	echo "-- Migration: $(name)" > $$filename; \
	echo "-- Rollback: $(name)" > $$down_filename; \
	echo "Created migration files: $$filename and $$down_filename"

db-seed: db-migrate
	@if [ "$(ENV)" = "development" ]; then \
		PGPASSWORD=$${DB_PASSWORD} psql -h $${DB_HOST} -p $${DB_PORT} -U $${DB_USER} -d $${DB_NAME} -f seeds/01_dev_seed.sql; \
	else \
		echo "Seeding is only available in development environment"; \
		exit 1; \
	fi

db-reset: db-down
	docker compose rm -f db
	docker volume rm cashone_postgres_data || true
	$(MAKE) db-up

db-shell:
	@PGPASSWORD=$${DB_PASSWORD} psql -h $${DB_HOST} -p $${DB_PORT} -U $${DB_USER} -d $${DB_NAME}

db-logs:
	docker compose logs -f db

db-test:
	docker compose up -d
	cd $(MIGRATE_PATH) && $(GORUN) main.go -command up
	cd $(MIGRATE_PATH) && $(GORUN) main.go -command down
	docker compose stop test

# Development helpers
dev-reset: clean db-reset
	$(MAKE) dev

# CI targets
ci: check build docs db-test

# Release targets
release:
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Use: make release VERSION=X.Y.Z"; \
		exit 1; \
	fi
	./scripts/release.sh $(VERSION)

# Help target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Setup and initialization:"
	@echo "  init        - Initialize environment (create .env file and directories)"
	@echo "  setup       - Initialize environment and install dependencies"
	@echo ""
	@echo "Development workflow:"
	@echo "  dev-live    - Run application with live reload (recommended for development)"
	@echo "  dev-install - Install development tools (air, swag, golangci-lint)"
	@echo "  dev-reset   - Reset development environment and start with live reload"
	@echo "  dev         - Setup and run application (without live reload)"
	@echo ""
	@echo "Version management:"
	@echo "  version     - Display version information"
	@echo "  release     - Create a new release (requires VERSION=X.Y.Z)"
	@echo ""
	@echo "Documentation:"
	@echo "  docs        - Generate API documentation"
	@echo "  serve-docs  - Generate and serve API documentation"
	@echo ""
	@echo "Testing and quality:"
	@echo "  lint            - Run linter"
	@echo "  check           - Run all checks (lint + test-coverage)"
	@echo ""
	@echo "Database operations:"
	@echo "  db-up           - Start database container and run migrations"
	@echo "  db-down         - Stop database container"
	@echo "  db-migrate      - Run database migrations"
	@echo "  db-rollback     - Rollback all migrations"
	@echo "  db-status       - Show migration status"
	@echo "  db-new          - Create new migration (use: make db-new name=add_something)"
	@echo "  db-seed         - Seed database with development data"
	@echo "  db-reset        - Reset database (delete and recreate)"
	@echo "  db-shell        - Open database shell"
	@echo "  db-logs         - Show database container logs"
	@echo "  db-test         - Run migration tests"
	@echo ""
	@echo "Docker operations:"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-run      - Run application in Docker"
	@echo "  docker-down     - Stop Docker containers"
	@echo ""
	@echo "Build operations:"
	@echo "  build           - Build the application"
	@echo "  build-linux     - Build for Linux"
	@echo "  clean           - Clean build files"
	@echo ""
	@echo "CI/CD:"
	@echo "  ci              - Run CI checks (lint + test-coverage + build + docs + db-test)"
