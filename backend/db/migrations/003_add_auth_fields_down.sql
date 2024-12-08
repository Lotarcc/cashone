-- Drop refresh_tokens table and its trigger
DROP TRIGGER IF EXISTS update_refresh_tokens_updated_at ON refresh_tokens;
DROP TABLE IF EXISTS refresh_tokens;

-- Remove authentication-related fields from users table
ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified,
    DROP COLUMN IF EXISTS last_login_at;
