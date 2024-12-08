-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Set up proper encoding and locale
SET client_encoding = 'UTF8';

-- Create schema and grant permissions
CREATE SCHEMA IF NOT EXISTS public;
GRANT ALL ON SCHEMA public TO PUBLIC;

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
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO PUBLIC;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO PUBLIC;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON FUNCTIONS TO PUBLIC;

-- Set up basic security policies
ALTER DATABASE postgres SET session_preload_libraries = 'auto_explain';
ALTER DATABASE postgres SET auto_explain.log_min_duration = '3s';

-- Vacuum analyze to update statistics
ANALYZE VERBOSE;
