-- Create ledger lines table
CREATE TABLE IF NOT EXISTS ledger_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    debit_amount BIGINT NOT NULL DEFAULT 0,  -- Amount in paise (0 if credit)
    credit_amount BIGINT NOT NULL DEFAULT 0, -- Amount in paise (0 if debit)
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Ensure line has either debit or credit, not both
    CONSTRAINT ledger_lines_amount_check CHECK (
        (debit_amount > 0 AND credit_amount = 0) OR
        (credit_amount > 0 AND debit_amount = 0)
    )
);

-- Create indexes
CREATE INDEX idx_ledger_lines_entry_id ON ledger_lines(entry_id);
CREATE INDEX idx_ledger_lines_account_id ON ledger_lines(account_id);
CREATE INDEX idx_ledger_lines_created_at ON ledger_lines(created_at DESC);

-- Function to update account balances when a journal entry is posted
CREATE OR REPLACE FUNCTION update_account_balances()
RETURNS TRIGGER AS $$
DECLARE
    line RECORD;
    account RECORD;
    new_balance BIGINT;
BEGIN
    -- Only process when entry status changes to 'posted'
    IF NEW.status = 'posted' AND (OLD.status IS NULL OR OLD.status != 'posted') THEN
        -- Loop through all lines in this entry
        FOR line IN
            SELECT account_id, debit_amount, credit_amount
            FROM ledger_lines
            WHERE entry_id = NEW.id
        LOOP
            -- Get account info
            SELECT id, type, balance, debit_total, credit_total
            INTO account
            FROM accounts
            WHERE id = line.account_id
            FOR UPDATE; -- Lock row for update

            -- Update debit/credit totals
            UPDATE accounts
            SET
                debit_total = debit_total + line.debit_amount,
                credit_total = credit_total + line.credit_amount
            WHERE id = account.id;

            -- Calculate new balance based on account type
            IF account.type IN ('asset', 'expense') THEN
                -- Debit normal accounts: debit increases, credit decreases
                new_balance := account.balance + line.debit_amount - line.credit_amount;
            ELSE
                -- Credit normal accounts: credit increases, debit decreases
                new_balance := account.balance + line.credit_amount - line.debit_amount;
            END IF;

            -- Update balance
            UPDATE accounts
            SET balance = new_balance
            WHERE id = account.id;
        END LOOP;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update account balances when entry is posted
CREATE TRIGGER update_account_balances_on_post
    AFTER UPDATE ON journal_entries
    FOR EACH ROW
    WHEN (NEW.status = 'posted')
    EXECUTE FUNCTION update_account_balances();

-- Function to validate journal entry is balanced before posting
CREATE OR REPLACE FUNCTION validate_journal_entry_balanced()
RETURNS TRIGGER AS $$
DECLARE
    total_debits BIGINT;
    total_credits BIGINT;
BEGIN
    -- Only validate when posting
    IF NEW.status = 'posted' AND (OLD.status IS NULL OR OLD.status != 'posted') THEN
        -- Calculate totals
        SELECT
            COALESCE(SUM(debit_amount), 0),
            COALESCE(SUM(credit_amount), 0)
        INTO total_debits, total_credits
        FROM ledger_lines
        WHERE entry_id = NEW.id;

        -- Check if balanced
        IF total_debits != total_credits THEN
            RAISE EXCEPTION 'Journal entry % is not balanced: debits=%, credits=%',
                NEW.entry_number, total_debits, total_credits;
        END IF;

        -- Check if entry has at least 2 lines
        IF (SELECT COUNT(*) FROM ledger_lines WHERE entry_id = NEW.id) < 2 THEN
            RAISE EXCEPTION 'Journal entry % must have at least 2 lines', NEW.entry_number;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to validate entry is balanced before posting
CREATE TRIGGER validate_entry_balanced_on_post
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    WHEN (NEW.status = 'posted')
    EXECUTE FUNCTION validate_journal_entry_balanced();

-- Create view for account balances with account details
CREATE OR REPLACE VIEW account_balances AS
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
    -- Calculate if balance is normal or abnormal
    CASE
        WHEN a.type IN ('asset', 'expense') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('asset', 'expense') AND a.balance < 0 THEN 'abnormal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance >= 0 THEN 'normal'
        WHEN a.type IN ('liability', 'equity', 'revenue') AND a.balance < 0 THEN 'abnormal'
    END AS balance_status
FROM accounts a;

-- Create view for general ledger (all posted transactions)
CREATE OR REPLACE VIEW general_ledger AS
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
