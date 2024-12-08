-- Rollback initial schema

-- Drop triggers
DROP TRIGGER IF EXISTS update_monobank_integrations_updated_at ON monobank_integrations;
DROP TRIGGER IF EXISTS update_transactions_updated_at ON transactions;
DROP TRIGGER IF EXISTS update_cards_updated_at ON cards;
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_transactions_monobank_id;
DROP INDEX IF EXISTS idx_transactions_transaction_date;
DROP INDEX IF EXISTS idx_transactions_category_id;
DROP INDEX IF EXISTS idx_transactions_card_id;
DROP INDEX IF EXISTS idx_transactions_user_id;
DROP INDEX IF EXISTS idx_cards_monobank_account_id;
DROP INDEX IF EXISTS idx_cards_user_id;
DROP INDEX IF EXISTS idx_categories_user_id;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables in correct order to handle foreign key constraints
DROP TABLE IF EXISTS monobank_integrations;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS cards;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS users;

-- Drop extensions if they're no longer needed
-- Note: Be careful with dropping extensions as they might be used by other databases
-- DROP EXTENSION IF EXISTS "uuid-ossp";
