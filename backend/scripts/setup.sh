#!/bin/bash
set -e

echo "Setting up CashOne backend..."

# Navigate to app directory
cd "$(dirname "$0")/../app"

# Download dependencies
echo "Downloading dependencies..."
go mod download
go mod verify

# Run tests
echo "Running tests..."
go test -v ./...

# Build the application
echo "Building application..."
go build -o ../bin/cashone ./cmd

echo "Setup completed successfully!"
