# Expense Tracker Application - Phase 1: Setup and Backend Development (2 weeks)

This phase focuses on setting up the development environment and building the backend API using Go and PostgreSQL.  The Monobank API will be integrated to fetch transaction data.

**1. Set up development environment (2 days):**

*   Install Go: Download and install the latest stable version of Go from the official website.  Verify the installation by running `go version`.
*   Install PostgreSQL: Download and install PostgreSQL from the official website.  Create a new database named `expense_tracker`.
*   Install Node.js and npm: Download and install Node.js and npm (Node Package Manager) from the official website. Verify the installation by running `node -v` and `npm -v`.
*   Clone the `vtopc/go-monobank` library: Use `go get github.com/vtopc/go-monobank` to download the necessary library.
*   Install Docker: Download and install Docker from the official website. Verify the installation by running `docker --version`.

**1.1 Containerization Setup (1 day):**

*   **PostgreSQL Container:**
    - Create a Dockerfile for PostgreSQL with necessary configurations
    - Set up environment variables for database credentials
    - Configure persistent volume for data storage
    - Set up initialization scripts for database schema

*   **Application Container:**
    - Create a multi-stage Dockerfile for the Go application
    - First stage for building the Go binary
    - Second stage for running the application with minimal base image
    - Configure environment variables for database connection and API keys
    - Set up proper logging and monitoring

*   **Docker Compose:**
    - Create docker-compose.yml to orchestrate both containers
    - Configure networking between containers
    - Set up volume mappings for persistence
    - Define environment variables
    - Configure health checks for both services

**2. Create backend API (10 days):**

*   **User Authentication (2 days):**  Implement user authentication using a suitable Go library (e.g., `gorilla/mux` for routing and a library for authentication like `github.com/google/uuid` for generating unique user IDs).  Consider using JWT (JSON Web Tokens) for secure authentication.  This will include endpoints for user registration and login.
*   **Monobank Transaction Fetching (2 days):** Integrate the `vtopc/go-monobank` library to fetch transaction data from the Monobank API.  Handle API rate limits and potential errors gracefully.  Implement proper error handling and logging.
*   **Transaction Data Storage (2 days):** Design and implement database interactions using Go's database/sql package and PostgreSQL.  This includes creating tables for users, transactions, and categories.  Ensure data integrity and consistency.
*   **Expense/Income/Transfer Entry (2 days):** Create API endpoints for adding, updating, and deleting expense, income, and transfer records.  Validate input data to prevent errors.
*   **Category Management (2 days):** Implement API endpoints for managing categories (adding, updating, deleting).  Consider using a hierarchical category structure for better organization.

**3. Database Design (1 day):**

*   Design a PostgreSQL schema for users, transactions, and categories.  Consider using appropriate data types and constraints to ensure data integrity.  Create SQL scripts for creating the database tables.

**4. Testing (2 days):**

*   Implement unit tests for individual API functions using a testing framework like `testing` (built into Go).
*   Implement integration tests to verify the interaction between different components of the backend API.  Use a testing framework like `testify/assert` for assertions.
*   Add container-specific tests to verify proper containerization and inter-container communication.
