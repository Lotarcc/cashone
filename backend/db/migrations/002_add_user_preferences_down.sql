-- Enable warnings
SET client_min_messages TO WARNING;

-- Start transaction
BEGIN;

-- Drop dependent objects first
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_preferences_lookup;
DROP INDEX IF EXISTS idx_user_preferences_category;
DROP INDEX IF EXISTS idx_user_preferences_user_id;

-- Drop the table
DROP TABLE IF EXISTS user_preferences;

-- End transaction
COMMIT;
