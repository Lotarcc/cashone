#!/bin/bash

# Load environment variables
source .env

# Database configuration for tests
TEST_DB_NAME="cashone_test"
TEST_DB_USER="cashone_user"
TEST_DB_PASSWORD="cashone_password"

# Create test database and user
PGPASSWORD=${CASHONE_DATABASE_PASSWORD} psql -h localhost -U ${CASHONE_DATABASE_USER} -d postgres -c "DROP DATABASE IF EXISTS ${TEST_DB_NAME};"
PGPASSWORD=${CASHONE_DATABASE_PASSWORD} psql -h localhost -U ${CASHONE_DATABASE_USER} -d postgres -c "CREATE DATABASE ${TEST_DB_NAME};"

# Grant privileges
PGPASSWORD=${CASHONE_DATABASE_PASSWORD} psql -h localhost -U ${CASHONE_DATABASE_USER} -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${TEST_DB_NAME} TO ${TEST_DB_USER};"

echo "Test database setup complete"
