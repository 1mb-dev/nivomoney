-- Drop triggers
DROP TRIGGER IF EXISTS sync_wallet_available_balance_trigger ON wallets;
DROP TRIGGER IF EXISTS update_wallets_updated_at ON wallets;

-- Drop functions
DROP FUNCTION IF EXISTS sync_wallet_available_balance();

-- Drop table
DROP TABLE IF EXISTS wallets;
