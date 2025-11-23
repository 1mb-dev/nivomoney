-- Create accounts table (Chart of Accounts)
CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'INR',
    parent_id UUID REFERENCES accounts(id) ON DELETE RESTRICT,
    balance BIGINT NOT NULL DEFAULT 0, -- Balance in smallest unit (paise for INR)
    debit_total BIGINT NOT NULL DEFAULT 0,  -- Lifetime debit total
    credit_total BIGINT NOT NULL DEFAULT 0, -- Lifetime credit total
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT accounts_type_check CHECK (type IN ('asset', 'liability', 'equity', 'revenue', 'expense')),
    CONSTRAINT accounts_status_check CHECK (status IN ('active', 'inactive', 'closed')),
    CONSTRAINT accounts_currency_check CHECK (currency ~* '^[A-Z]{3}$')
);

-- Create indexes
CREATE INDEX idx_accounts_code ON accounts(code);
CREATE INDEX idx_accounts_type ON accounts(type);
CREATE INDEX idx_accounts_parent_id ON accounts(parent_id);
CREATE INDEX idx_accounts_status ON accounts(status);

-- Create trigger to update updated_at
CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Insert standard chart of accounts for Indian neobank

-- Assets (1000-1999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('1000', 'Cash and Bank Accounts', 'asset', 'INR', 'active'),
('1100', 'Accounts Receivable', 'asset', 'INR', 'active'),
('1200', 'Loans Receivable', 'asset', 'INR', 'active'),
('1300', 'Investments', 'asset', 'INR', 'active'),
('1400', 'Fixed Assets', 'asset', 'INR', 'active'),
('1500', 'Other Assets', 'asset', 'INR', 'active');

-- Liabilities (2000-2999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('2000', 'Accounts Payable', 'liability', 'INR', 'active'),
('2100', 'Customer Deposits', 'liability', 'INR', 'active'),
('2200', 'Borrowings', 'liability', 'INR', 'active'),
('2300', 'Taxes Payable', 'liability', 'INR', 'active'),
('2400', 'Other Liabilities', 'liability', 'INR', 'active');

-- Equity (3000-3999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('3000', 'Share Capital', 'equity', 'INR', 'active'),
('3100', 'Retained Earnings', 'equity', 'INR', 'active'),
('3200', 'Reserves', 'equity', 'INR', 'active');

-- Revenue (4000-4999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('4000', 'Interest Income', 'revenue', 'INR', 'active'),
('4100', 'Fee Income', 'revenue', 'INR', 'active'),
('4200', 'Transaction Fees', 'revenue', 'INR', 'active'),
('4300', 'Other Income', 'revenue', 'INR', 'active');

-- Expenses (5000-5999)
INSERT INTO accounts (code, name, type, currency, status) VALUES
('5000', 'Interest Expense', 'expense', 'INR', 'active'),
('5100', 'Operating Expenses', 'expense', 'INR', 'active'),
('5200', 'Salary and Wages', 'expense', 'INR', 'active'),
('5300', 'Technology Expenses', 'expense', 'INR', 'active'),
('5400', 'Marketing Expenses', 'expense', 'INR', 'active'),
('5500', 'Compliance and Legal', 'expense', 'INR', 'active'),
('5600', 'Other Expenses', 'expense', 'INR', 'active');
