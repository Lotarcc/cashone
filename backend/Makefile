# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=cashone
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=./app/cmd

# Version information
VERSION ?= $(shell git describe --tags --always --dirty || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d %H:%M:%S UTC')

.PHONY: all init setup build run test test-coverage clean build-linux docker-build docker-run docker-down db-migrate db-seed docs version

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

# Documentation targets
docs:
	./scripts/generate-docs.sh

serve-docs: docs
	@echo "Opening Swagger UI..."
	@which xdg-open > /dev/null && xdg-open http://localhost:8080/swagger/index.html || \
	which open > /dev/null && open http://localhost:8080/swagger/index.html || \
	echo "Please open http://localhost:8080/swagger/index.html in your browser"

# Build targets
build: update-version
	cd $(MAIN_PATH) && $(GOBUILD) -o ../../bin/$(BINARY_NAME) -v

build-linux: update-version
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) -v

# Run targets
run:
	cd $(MAIN_PATH) && $(GORUN) main.go

# Test and quality targets
test:
	cd app && $(GOTEST) -v ./...

test-coverage:
	./scripts/test-coverage.sh

test-watch:
	cd app && $(GOTEST) -v ./... -count=1 -watch

test-short:
	cd app && $(GOTEST) -v ./... -short

lint:
	cd app && golangci-lint run

check: lint test-coverage

# Clean targets
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)
	rm -rf coverage/
	rm -rf docs/

# Docker targets
docker-build: update-version
	docker build -t $(BINARY_NAME) .

docker-run:
	docker-compose up

docker-down:
	docker-compose down

# Database targets
db-up:
	docker-compose up -d db
	@echo "Waiting for database to be ready..."
	@sleep 5
	$(MAKE) db-migrate

db-down:
	docker-compose stop db

db-migrate:
	./scripts/db-migrate.sh

db-seed: db-migrate
	@if [ "$(ENV)" = "development" ]; then \
		PGPASSWORD=$${DB_PASSWORD} psql -h $${DB_HOST} -p $${DB_PORT} -U $${DB_USER} -d $${DB_NAME} -f seeds/01_dev_seed.sql; \
	else \
		echo "Seeding is only available in development environment"; \
		exit 1; \
	fi

db-reset: db-down
	docker-compose rm -f db
	docker volume rm cashone_postgres_data || true
	$(MAKE) db-up

# Development helpers
dev-reset: clean db-reset
	$(MAKE) dev

# CI targets
ci: check build docs

# Release targets
release:
	@if [ "$(VERSION)" = "" ]; then \
		echo "Error: VERSION is required. Use: make release VERSION=X.Y.Z"; \
		exit 1; \
	fi
	@echo "Creating release $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	$(MAKE) update-version

# Help target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Setup and initialization:"
	@echo "  init        - Initialize environment (create .env file and directories)"
	@echo "  setup       - Initialize environment and install dependencies"
	@echo ""
	@echo "Version management:"
	@echo "  version     - Display version information"
	@echo "  release     - Create a new release (requires VERSION=X.Y.Z)"
	@echo ""
	@echo "Development workflow:"
	@echo "  dev         - Setup and run application in development mode"
	@echo "  dev-reset   - Reset development environment"
	@echo "  run         - Run the application locally"
	@echo ""
	@echo "Documentation:"
	@echo "  docs        - Generate API documentation"
	@echo "  serve-docs  - Generate and serve API documentation"
	@echo ""
	@echo "Testing and quality:"
	@echo "  test            - Run all tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-watch      - Run tests in watch mode"
	@echo "  test-short      - Run only short tests"
	@echo "  lint            - Run linter"
	@echo "  check           - Run all checks (lint + test-coverage)"
	@echo ""
	@echo "Database operations:"
	@echo "  db-up           - Start database container and run migrations"
	@echo "  db-down         - Stop database container"
	@echo "  db-migrate      - Run database migrations"
	@echo "  db-seed         - Seed database with development data"
	@echo "  db-reset        - Reset database (delete and recreate)"
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
	@echo "  ci              - Run CI checks (lint + test-coverage + build + docs)"
