-- Migration template
-- Replace XXX with the next migration number (e.g., 002, 003)
-- Replace description with a brief description of what this migration does
-- Example filename: 002_add_user_settings.sql

-- Enable warnings
SET client_min_messages TO WARNING;

-- Start transaction
BEGIN;

-- Your migration SQL here
-- Example:
/*
CREATE TABLE example_table (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes
CREATE INDEX idx_example_table_user_id ON example_table(user_id);
CREATE INDEX idx_example_table_status ON example_table(status);
CREATE INDEX idx_example_table_created_at ON example_table(created_at);

-- Add updated_at trigger
CREATE TRIGGER update_example_table_updated_at
    BEFORE UPDATE ON example_table
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add any necessary foreign key constraints
ALTER TABLE example_table
    ADD CONSTRAINT fk_example_table_user
    FOREIGN KEY (user_id)
    REFERENCES users(id)
    ON DELETE CASCADE;

-- Add any necessary unique constraints
ALTER TABLE example_table
    ADD CONSTRAINT uq_example_table_name
    UNIQUE (user_id, name);

-- Add any necessary check constraints
ALTER TABLE example_table
    ADD CONSTRAINT chk_example_table_status
    CHECK (status IN ('active', 'inactive', 'pending'));

-- Add any necessary default values or data
INSERT INTO example_table (user_id, name, description)
SELECT 
    id as user_id,
    'Default Settings' as name,
    'Auto-generated default settings' as description
FROM users;
*/

-- End transaction
COMMIT;
