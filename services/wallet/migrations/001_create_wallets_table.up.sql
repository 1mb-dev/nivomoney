-- Create wallets table
CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,  -- Reference to identity service users table
    type VARCHAR(20) NOT NULL DEFAULT 'savings',
    currency CHAR(3) NOT NULL DEFAULT 'INR',
    balance BIGINT NOT NULL DEFAULT 0,  -- In smallest unit (paise for INR)
    available_balance BIGINT NOT NULL DEFAULT 0,  -- Balance minus holds
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    ledger_account_id UUID NOT NULL,  -- Reference to ledger service account
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    closed_reason TEXT,

    CONSTRAINT wallets_type_check CHECK (type IN ('savings', 'current', 'fixed')),
    CONSTRAINT wallets_status_check CHECK (status IN ('active', 'frozen', 'closed', 'inactive')),
    CONSTRAINT wallets_balance_check CHECK (balance >= 0),
    CONSTRAINT wallets_available_balance_check CHECK (available_balance >= 0 AND available_balance <= balance),
    CONSTRAINT wallets_closed_check CHECK (
        (status = 'closed' AND closed_at IS NOT NULL AND closed_reason IS NOT NULL) OR
        (status != 'closed' AND closed_at IS NULL AND closed_reason IS NULL)
    )
);

-- Create indexes
CREATE INDEX idx_wallets_user_id ON wallets(user_id);
CREATE INDEX idx_wallets_status ON wallets(status);
CREATE INDEX idx_wallets_ledger_account ON wallets(ledger_account_id);
CREATE INDEX idx_wallets_created_at ON wallets(created_at DESC);

-- Create unique constraint: one active wallet per user per type per currency
CREATE UNIQUE INDEX idx_wallets_unique_active
    ON wallets(user_id, type, currency)
    WHERE status IN ('active', 'frozen', 'inactive');

-- Create trigger to update updated_at
CREATE TRIGGER update_wallets_updated_at
    BEFORE UPDATE ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to sync balance with available_balance on insert
CREATE OR REPLACE FUNCTION sync_wallet_available_balance()
RETURNS TRIGGER AS $$
BEGIN
    -- On insert, set available_balance same as balance
    IF TG_OP = 'INSERT' THEN
        NEW.available_balance := NEW.balance;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER sync_wallet_available_balance_trigger
    BEFORE INSERT ON wallets
    FOR EACH ROW
    EXECUTE FUNCTION sync_wallet_available_balance();
