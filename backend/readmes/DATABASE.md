# Database Management

This document describes the database management system used in the CashOne project.

## Overview

The project uses PostgreSQL as its primary database and includes a robust migration system for managing database schema changes. The system supports:

- Versioned migrations
- Rollback capabilities
- Development seeding
- Status checking
- Transaction-safe migrations
- Automated migration testing
- Migration templates and generators

## Migration Files

Migrations are stored in the `migrations` directory and follow this naming convention:

```
NNN_description.sql        # Up migration
NNN_description_down.sql   # Down migration (rollback)
```

For example:
- `001_initial_schema.sql`
- `001_initial_schema_down.sql`

## Commands

The following make commands are available for database management:

```bash
# Create a new migration
make db-new name=add_user_settings

# Start database and run migrations
make db-up

# Stop database
make db-down

# Run migrations
make db-migrate

# Rollback last migration
make db-rollback

# Show migration status
make db-status

# Seed development data
make db-seed

# Reset database (delete and recreate)
make db-reset

# Open database shell
make db-shell

# Show database logs
make db-logs

# Run migration tests
make db-test
```

## Creating New Migrations

### Using the Migration Generator

The easiest way to create a new migration is to use the generator:

```bash
make db-new name=add_user_settings
```

This will:
1. Generate the next migration number automatically
2. Create both up and down migration files
3. Use the migration templates
4. Open the files in your editor (if configured)

The generator creates:
- `migrations/XXX_add_user_settings.sql` (up migration)
- `migrations/XXX_add_user_settings_down.sql` (down migration)

### Migration Templates

The project includes templates for both up and down migrations:
- `migrations/template_migration.sql`
- `migrations/template_migration_down.sql`

These templates include:
- Best practices for migration structure
- Common SQL patterns
- Transaction handling
- Error handling
- Documentation examples
- Index creation patterns
- Constraint examples

### Manual Creation

If you prefer to create migrations manually:

1. Copy the template files:
   ```bash
   cp migrations/template_migration.sql migrations/002_add_user_settings.sql
   cp migrations/template_migration_down.sql migrations/002_add_user_settings_down.sql
   ```

2. Edit the files following the template patterns:
   ```sql
   -- 002_add_user_settings.sql
   BEGIN;

   CREATE TABLE user_settings (
       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
       user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       key VARCHAR(255) NOT NULL,
       value TEXT,
       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
       UNIQUE(user_id, key)
   );

   -- Add trigger for updated_at
   CREATE TRIGGER update_user_settings_updated_at
       BEFORE UPDATE ON user_settings
       FOR EACH ROW
       EXECUTE FUNCTION update_updated_at_column();

   COMMIT;
   ```

## Testing Migrations

The project includes automated testing for migrations. To run migration tests:

```bash
make db-test
```

This will:
1. Create a test database
2. Run all migrations
3. Verify table creation
4. Test rollback functionality
5. Re-apply migrations
6. Test development seeds
7. Clean up the test database

The test script (`scripts/test-migrations.sh`) checks:
- Migration application
- Migration rollback
- Table creation/deletion
- Trigger creation
- Seed data loading
- Error handling

Always run migration tests before committing new migrations:
```bash
# Full test sequence
make db-test

# If successful, commit your changes
git add migrations/
git commit -m "feat: add new migration for user settings"
```

## Development Seeds

Development seed data is stored in the `seeds` directory and is automatically loaded in development environment after migrations. To add new seed data:

1. Create a new seed file in the `seeds` directory
2. Add your SQL INSERT statements
3. Run `make db-seed` to apply the new seeds

## Best Practices

1. **Use the generator**: Always use `make db-new` to create migrations.

2. **Follow the templates**: The templates include best practices and patterns.

3. **Always create down migrations**: This ensures you can rollback changes if needed.

4. **Use transactions**: All migrations should be wrapped in transactions.

5. **Test thoroughly**: Always run `make db-test` before committing migrations.

6. **Keep migrations small**: Each migration should make a focused set of changes.

7. **Use meaningful names**: Migration names should clearly describe their purpose.

8. **Include indexes**: Add appropriate indexes in the same migration that creates the table.

9. **Foreign key constraints**: Always include appropriate foreign key constraints to maintain data integrity.

10. **Test both directions**: Ensure both up and down migrations work correctly.

## Troubleshooting

### Common Issues

1. **Migration fails to apply**
   - Check the error message in the logs
   - Verify the SQL syntax
   - Ensure dependencies (tables, functions, etc.) exist
   - Try running `make db-status` to see the current state

2. **Rollback fails**
   - Check if the down migration properly reverts all changes
   - Verify the order of operations (drop dependent objects first)

3. **Seed data fails to load**
   - Verify the data matches the current schema
   - Check for any unique constraint violations
   - Ensure referenced data exists

4. **Migration tests fail**
   - Check the test output for specific failures
   - Verify both up and down migrations
   - Ensure all tables are properly created/dropped
   - Check if seeds are valid for the current schema

### Recovery Steps

1. If a migration fails:
   ```bash
   # Check the status
   make db-status
   
   # Rollback if needed
   make db-rollback
   
   # Fix the migration file
   
   # Try again
   make db-migrate
   ```

2. If the database is in a bad state:
   ```bash
   # Reset everything
   make db-reset
   ```

3. If tests are failing:
   ```bash
   # Run with detailed output
   ENV=development ./scripts/test-migrations.sh
   
   # Check the logs
   make db-logs
   ```

## Schema Documentation

The current database schema includes:

- `users`: User accounts and authentication
- `categories`: Transaction categories
- `cards`: Bank cards (both manual and Monobank)
- `transactions`: Financial transactions
- `monobank_integrations`: Monobank API integration data

For detailed schema information, see the migration files in the `migrations` directory.
