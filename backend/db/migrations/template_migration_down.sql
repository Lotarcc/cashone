-- Down migration template
-- Replace XXX with the migration number matching your up migration
-- Replace description with the same description as your up migration
-- Example filename: 002_add_user_settings_down.sql

-- Enable warnings
SET client_min_messages TO WARNING;

-- Start transaction
BEGIN;

-- Your down migration SQL here
-- IMPORTANT: Drop objects in reverse order of creation
-- Example:
/*
-- Drop dependent objects first (foreign keys, triggers, etc.)
DROP TRIGGER IF EXISTS update_example_table_updated_at ON example_table;

-- Drop indexes
DROP INDEX IF EXISTS idx_example_table_user_id;
DROP INDEX IF EXISTS idx_example_table_status;
DROP INDEX IF EXISTS idx_example_table_created_at;

-- Drop constraints (if they were added separately)
ALTER TABLE example_table
    DROP CONSTRAINT IF EXISTS chk_example_table_status;

ALTER TABLE example_table
    DROP CONSTRAINT IF EXISTS uq_example_table_name;

ALTER TABLE example_table
    DROP CONSTRAINT IF EXISTS fk_example_table_user;

-- Finally, drop the table
DROP TABLE IF EXISTS example_table;

-- If you added any types, functions, or other objects, drop them too
-- DROP TYPE IF EXISTS example_status;
-- DROP FUNCTION IF EXISTS example_function();
*/

-- End transaction
COMMIT;

-- Notes:
-- 1. Always use IF EXISTS to avoid errors if objects don't exist
-- 2. Drop dependent objects before their parents
-- 3. Drop in reverse order of creation
-- 4. Include all objects that were created in the up migration
-- 5. Test the down migration after creating it
