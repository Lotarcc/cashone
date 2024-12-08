-- Enable warnings
SET client_min_messages TO WARNING;

-- Start transaction
BEGIN;

-- Create user preferences table
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL,
    key VARCHAR(255) NOT NULL,
    value JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, category, key)
);

-- Add indexes
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_preferences_category ON user_preferences(category);
CREATE INDEX idx_user_preferences_lookup ON user_preferences(user_id, category);

-- Add updated_at trigger
CREATE TRIGGER update_user_preferences_updated_at
    BEFORE UPDATE ON user_preferences
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add default preferences for existing users
INSERT INTO user_preferences (user_id, category, key, value)
SELECT 
    id as user_id,
    'display' as category,
    'theme' as key,
    '{"mode": "light", "color": "blue"}'::jsonb as value
FROM users;

INSERT INTO user_preferences (user_id, category, key, value)
SELECT 
    id as user_id,
    'notifications' as category,
    'settings' as key,
    '{"email": true, "push": true}'::jsonb as value
FROM users;

-- End transaction
COMMIT;
