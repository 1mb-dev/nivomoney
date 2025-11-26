-- Revert account code length back to 20 characters
ALTER TABLE accounts ALTER COLUMN code TYPE VARCHAR(20);
