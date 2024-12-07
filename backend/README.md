# CashOne Backend

A Go-based backend service for the CashOne expense tracking application. This service provides REST APIs for managing personal finances, including manual and Monobank-integrated transactions.

## Project Structure

```
backend/
├── app/
│   ├── cmd/                    # Application entry point
│   ├── config/                 # Configuration files
│   ├── domain/                 # Domain layer (entities, interfaces)
│   │   ├── entity/            # Domain entities
│   │   ├── repository/        # Repository interfaces
│   │   └── service/           # Service interfaces
│   └── infrastructure/         # Infrastructure layer
│       ├── database/          # Database connection and migrations
│       ├── handler/           # HTTP handlers
│       ├── repository/        # Repository implementations
│       └── service/           # Service implementations
├── docker/                     # Docker-related files
│   └── postgres/              # PostgreSQL configuration
├── scripts/                    # Development and deployment scripts
├── bin/                        # Compiled binaries
├── Dockerfile                  # Backend service Dockerfile
└── Makefile                   # Build and development commands
```

## Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- PostgreSQL 15 or later
- Make
- golangci-lint (for development)

## Quick Start

1. Clone the repository
2. Initialize the environment:
   ```bash
   make init
   ```
3. Install dependencies and set up the project:
   ```bash
   make setup
   ```
4. Start the development environment:
   ```bash
   make dev
   ```

The application will be available at http://localhost:8080

## Development

### Environment Setup

The project uses environment variables for configuration. To set up your environment:

1. Run `make init` to create a `.env` file from `.env.example`
2. Edit the `.env` file to match your local setup
3. Run `make setup` to install dependencies and verify the setup

### Available Make Commands

```bash
# Initialize environment
make init

# Setup project
make setup

# Development mode
make dev

# Build the application
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean

# Start database only
make db-up

# Stop database
make db-down

# Run in Docker
make docker-run

# Stop Docker containers
make docker-down
```

For a complete list of available commands, run:
```bash
make help
```

### Testing

Run all tests:
```bash
make test
```

Run specific tests:
```bash
cd app && go test ./... -run TestName
```

### Code Quality

The project uses golangci-lint for code quality checks. Run:
```bash
make lint
```

## API Documentation

### Health Check
```
GET /health
Response: {"status": "ok"}
```

### Authentication
```
POST /api/v1/auth/register
POST /api/v1/auth/login
```

### Cards
```
GET    /api/v1/cards
POST   /api/v1/cards
GET    /api/v1/cards/:id
PUT    /api/v1/cards/:id
DELETE /api/v1/cards/:id
```

### Transactions
```
GET    /api/v1/transactions
POST   /api/v1/transactions
GET    /api/v1/transactions/:id
PUT    /api/v1/transactions/:id
DELETE /api/v1/transactions/:id
```

### Categories
```
GET    /api/v1/categories
POST   /api/v1/categories
GET    /api/v1/categories/:id
PUT    /api/v1/categories/:id
DELETE /api/v1/categories/:id
```

### Monobank Integration
```
POST   /api/v1/monobank/connect
GET    /api/v1/monobank/status
POST   /api/v1/monobank/sync
DELETE /api/v1/monobank/disconnect
POST   /api/v1/monobank/webhook
```

## Docker Support

The application can be run using Docker Compose, which sets up:
- Backend service
- PostgreSQL database

### Docker Commands

Start all services:
```bash
make docker-run
```

Stop all services:
```bash
make docker-down
```

Build backend image:
```bash
make docker-build
```

## Configuration

Configuration is handled through a combination of:
- Environment variables (`.env` file)
- Configuration files (`app/config/config.yaml`)

Important configuration files:
- `.env.example`: Template for environment variables
- `app/config/config.yaml`: Application configuration
- `docker-compose.yml`: Docker services configuration
- `.golangci.yml`: Linter configuration

## Contributing

1. Fork the repository
2. Create a new branch for your feature
3. Run tests and linter
4. Submit a pull request

### Development Workflow

1. Create a new branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and ensure:
   - All tests pass (`make test`)
   - Linter passes (`make lint`)
   - New features include tests
   - Documentation is updated

3. Commit your changes:
   ```bash
   git commit -m "feat: add your feature description"
   ```

4. Push to your fork and submit a pull request

## License

This project is licensed under the MIT License.
