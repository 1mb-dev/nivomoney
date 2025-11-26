-- Increase account code length from 20 to 50 characters to support wallet codes

-- Drop dependent views temporarily
DROP VIEW IF EXISTS account_balances CASCADE;
DROP VIEW IF EXISTS general_ledger CASCADE;

-- Alter the column type
ALTER TABLE accounts ALTER COLUMN code TYPE VARCHAR(50);

-- Recreate account_balances view
CREATE VIEW account_balances AS
SELECT
    a.id,
    a.code,
    a.name,
    a.type,
    a.currency,
    a.balance,
    a.debit_total,
    a.credit_total,
    a.status,
    a.created_at,
    a.updated_at,
    CASE
        WHEN a.type IN ('asset', 'expense') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('asset', 'expense') AND a.balance < 0 THEN 'abnormal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance < 0 THEN 'abnormal'
        ELSE NULL
    END AS balance_status
FROM accounts a;

-- Recreate general_ledger view
CREATE VIEW general_ledger AS
SELECT
    ll.id AS line_id,
    ll.entry_id,
    je.entry_number,
    je.type AS entry_type,
    je.description AS entry_description,
    je.posted_at,
    ll.account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    ll.debit_amount,
    ll.credit_amount,
    ll.description AS line_description,
    ll.created_at
FROM ledger_lines ll
JOIN journal_entries je ON ll.entry_id = je.id
JOIN accounts a ON ll.account_id = a.id
WHERE je.status = 'posted'
ORDER BY je.posted_at DESC, je.entry_number DESC, ll.created_at;
