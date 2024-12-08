# CashOne Backend

A Go-based backend service for the CashOne expense tracking application. This service provides REST APIs for managing personal finances, including manual and Monobank-integrated transactions.

## Project Structure

```
backend/
├── 
│   ├── cmd/                    # Application entry point
│   ├── config/                 # Configuration files
│   ├── domain/                 # Domain layer
│   │   ├── entity/            # Domain entities
│   │   ├── errors/            # Domain-specific errors
│   │   ├── repository/        # Repository interfaces
│   │   └── service/           # Service interfaces
│   ├── infrastructure/        # Infrastructure layer
│   │   ├── database/          # Database connection
│   │   ├── handler/           # HTTP handlers
│   │   ├── repository/        # Repository implementations
│   │   └── service/           # Service implementations
│   └── pkg/                   # Shared packages
│       ├── config/            # Configuration management
│       └── version/           # Version information
├── docker/                    # Docker configurations
│   └── postgres/              # PostgreSQL configuration
├── docs/                      # Documentation
│   └── DATABASE.md           # Database management guide
├── migrations/               # Database migrations
├── scripts/                  # Development and deployment scripts
├── seeds/                    # Development seed data
└── Makefile                 # Build and development commands
```

## Documentation

- [Database Management](docs/DATABASE.md)
- [API Documentation](http://localhost:8081/swagger/index.html) (when server is running)
- [Changelog](CHANGELOG.md)

## Prerequisites

- Go 1.23 or later
- Docker and Docker Compose
- PostgreSQL 15 or later
- Make

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
   make dev-live
   ```

The application will be available at http://localhost:8081

## Development

### Environment Setup

The project uses environment variables for configuration. To set up your environment:

1. Run `make init` to create a `.env` file from `.env.example`
2. Edit the `.env` file to match your local setup
3. Run `make setup` to install dependencies
4. Run `make dev-install` to install development tools

### Available Make Commands

```bash
# Setup and initialization
make init          # Initialize environment
make setup         # Install dependencies
make dev-install   # Install development tools

# Development workflow
make dev-live      # Run with live reload
make dev-reset     # Reset development environment
make run           # Run without live reload

# Database operations
make db-up         # Start database and run migrations
make db-migrate    # Run migrations
make db-rollback   # Rollback last migration
make db-status     # Show migration status
make db-seed       # Load development data
make db-reset      # Reset database
make db-shell      # Open database shell

# Testing and quality
make test          # Run tests
make test-coverage # Run with coverage
make lint          # Run linter
make check         # Run all checks

# Documentation
make docs          # Generate API docs
make serve-docs    # Serve API docs

# Docker operations
make docker-build  # Build image
make docker-run    # Run in Docker
make docker-down   # Stop containers
```

For a complete list of commands, run:
```bash
make help
```

### Testing

Run all tests:
```bash
make test
```

Run with coverage:
```bash
make test-coverage
```

### Code Quality

The project uses golangci-lint for code quality checks:
```bash
make lint
```

### Database Management

See [Database Management Guide](docs/DATABASE.md) for detailed information about:
- Migration system
- Creating new migrations
- Development seeds
- Best practices
- Troubleshooting

## API Documentation

When the server is running, Swagger documentation is available at:
```
http://localhost:8081/swagger/index.html
```

### Main Endpoints

- Authentication: `/api/v1/auth/*`
- Cards: `/api/v1/cards/*`
- Transactions: `/api/v1/transactions/*`
- Categories: `/api/v1/categories/*`
- Monobank Integration: `/api/v1/monobank/*`

## Docker Support

The application can be run using Docker Compose:
```bash
make docker-run
```

This sets up:
- Backend service
- PostgreSQL database

## Contributing

1. Fork the repository
2. Create a new branch
3. Make your changes
4. Run tests and linter
5. Submit a pull request

### Development Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature
   ```

2. Make changes and ensure:
   - Tests pass (`make test`)
   - Linter passes (`make lint`)
   - Documentation is updated
   - Database migrations are included if needed

3. Commit your changes:
   ```bash
   git commit -m "feat: add your feature"
   ```

4. Push and create a pull request

## License

This project is licensed under the MIT License.
