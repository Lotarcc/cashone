-- Development seed data

-- Insert test users
INSERT INTO users (email, password_hash, name) VALUES
    ('test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Test User'),  -- password: test123
    ('demo@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Demo User')   -- password: test123
ON CONFLICT (email) DO NOTHING;

-- Insert default categories
WITH user_ids AS (SELECT id FROM users WHERE email IN ('test@example.com', 'demo@example.com'))
INSERT INTO categories (name, type, user_id) 
SELECT category_name, category_type, user_id
FROM (
    VALUES
        ('Groceries', 'expense'),
        ('Rent', 'expense'),
        ('Transportation', 'expense'),
        ('Entertainment', 'expense'),
        ('Utilities', 'expense'),
        ('Healthcare', 'expense'),
        ('Shopping', 'expense'),
        ('Salary', 'income'),
        ('Freelance', 'income'),
        ('Investments', 'income'),
        ('Transfer In', 'transfer'),
        ('Transfer Out', 'transfer')
) AS c(category_name, category_type)
CROSS JOIN user_ids
ON CONFLICT DO NOTHING;

-- Insert test cards for each user
WITH user_data AS (
    SELECT id, email FROM users WHERE email IN ('test@example.com', 'demo@example.com')
)
INSERT INTO cards (user_id, card_name, masked_pan, balance, credit_limit, currency_code, is_manual)
SELECT 
    id,
    CASE 
        WHEN email = 'test@example.com' THEN 'Test Debit Card'
        ELSE 'Demo Credit Card'
    END,
    CASE 
        WHEN email = 'test@example.com' THEN '4111 11** **** 1111'
        ELSE '5555 55** **** 5555'
    END,
    CASE 
        WHEN email = 'test@example.com' THEN 1000000  -- $10,000.00
        ELSE 500000   -- $5,000.00
    END,
    CASE 
        WHEN email = 'test@example.com' THEN 0
        ELSE 1000000  -- $10,000.00 credit limit
    END,
    980, -- UAH
    true
FROM user_data
ON CONFLICT DO NOTHING;

-- Insert sample transactions
WITH user_data AS (
    SELECT u.id as user_id, c.id as card_id, cat.id as category_id
    FROM users u
    JOIN cards c ON c.user_id = u.id
    JOIN categories cat ON cat.user_id = u.id
    WHERE u.email = 'test@example.com'
    LIMIT 1
)
INSERT INTO transactions (
    user_id, card_id, category_id, amount, currency_code,
    type, description, transaction_date, balance_after
)
SELECT 
    user_id, card_id, category_id,
    amount, 980, -- UAH
    'expense',
    description,
    NOW() - (interval '1 day' * generate_series(0, 30)),
    1000000 - (amount * generate_series(0, 30))
FROM user_data,
(VALUES
    (5000, 'Coffee'),
    (15000, 'Lunch'),
    (100000, 'Groceries'),
    (500000, 'Rent')
) AS t(amount, description)
ON CONFLICT DO NOTHING;
