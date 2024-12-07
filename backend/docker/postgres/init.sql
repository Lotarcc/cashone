-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set up proper encoding and locale
SET client_encoding = 'UTF8';

-- Create application role with proper permissions
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = current_setting('POSTGRES_USER')) THEN
        CREATE ROLE current_setting('POSTGRES_USER') LOGIN PASSWORD current_setting('POSTGRES_PASSWORD');
    END IF;
END
$$;

-- Grant necessary permissions
ALTER ROLE current_setting('POSTGRES_USER') SET client_encoding TO 'utf8';
ALTER ROLE current_setting('POSTGRES_USER') SET default_transaction_isolation TO 'read committed';
ALTER ROLE current_setting('POSTGRES_USER') SET timezone TO 'UTC';

-- Create schema and grant permissions
CREATE SCHEMA IF NOT EXISTS public;
GRANT ALL ON SCHEMA public TO current_setting('POSTGRES_USER');
GRANT ALL ON ALL TABLES IN SCHEMA public TO current_setting('POSTGRES_USER');
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO current_setting('POSTGRES_USER');

-- Create custom functions

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Function to generate UUID v4
CREATE OR REPLACE FUNCTION generate_uuid_v4()
RETURNS uuid AS $$
BEGIN
    RETURN gen_random_uuid();
END;
$$ language 'plpgsql';

-- Set default privileges for future objects
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON TABLES TO current_setting('POSTGRES_USER');
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON SEQUENCES TO current_setting('POSTGRES_USER');
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON FUNCTIONS TO current_setting('POSTGRES_USER');

-- Create indexes on commonly used columns
CREATE INDEX IF NOT EXISTS idx_updated_at ON users(updated_at);
CREATE INDEX IF NOT EXISTS idx_created_at ON users(created_at);

-- Set up basic security policies
ALTER DATABASE current_setting('POSTGRES_DB')
    SET session_preload_libraries = 'auto_explain';

ALTER DATABASE current_setting('POSTGRES_DB')
    SET auto_explain.log_min_duration = '3s';

-- Vacuum analyze to update statistics
ANALYZE VERBOSE;
