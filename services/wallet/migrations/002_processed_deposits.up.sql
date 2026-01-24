-- ============================================================================
-- Processed Deposits Table for Idempotency
-- ============================================================================

CREATE TABLE IF NOT EXISTS processed_deposits (
    transaction_id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL,
    amount BIGINT NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_processed_deposits_processed_at ON processed_deposits(processed_at);
CREATE INDEX idx_processed_deposits_wallet ON processed_deposits(wallet_id);

COMMENT ON TABLE processed_deposits IS
'Tracks processed deposits to ensure idempotency. Prevents duplicate execution if transaction service retries.';

COMMENT ON COLUMN processed_deposits.transaction_id IS
'The transaction ID from the transaction service. Used as idempotency key.';
