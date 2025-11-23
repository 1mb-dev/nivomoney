-- Create KYC information table (India-specific)
CREATE TABLE IF NOT EXISTS user_kyc (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pan VARCHAR(10) NOT NULL,
    aadhaar VARCHAR(12) NOT NULL, -- Encrypted/hashed in production
    date_of_birth DATE NOT NULL,
    address JSONB NOT NULL,
    verified_at TIMESTAMP WITH TIME ZONE,
    rejected_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT kyc_status_check CHECK (status IN ('pending', 'verified', 'rejected', 'expired')),
    CONSTRAINT kyc_pan_check CHECK (pan ~* '^[A-Z]{5}[0-9]{4}[A-Z]$'),
    CONSTRAINT kyc_aadhaar_check CHECK (aadhaar ~* '^[2-9][0-9]{11}$'),
    CONSTRAINT kyc_pan_unique UNIQUE (pan),
    CONSTRAINT kyc_aadhaar_unique UNIQUE (aadhaar)
);

-- Create index on PAN for lookups
CREATE INDEX idx_kyc_pan ON user_kyc(pan);

-- Create index on status for filtering
CREATE INDEX idx_kyc_status ON user_kyc(status);

-- Create trigger to update updated_at
CREATE TRIGGER update_user_kyc_updated_at
    BEFORE UPDATE ON user_kyc
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Validate address JSONB structure (must have required fields)
CREATE OR REPLACE FUNCTION validate_address_jsonb()
RETURNS TRIGGER AS $$
BEGIN
    IF NOT (
        NEW.address ? 'street' AND
        NEW.address ? 'city' AND
        NEW.address ? 'state' AND
        NEW.address ? 'pin' AND
        NEW.address ? 'country'
    ) THEN
        RAISE EXCEPTION 'Address must contain street, city, state, pin, and country';
    END IF;

    -- Validate PIN code format (6 digits, cannot start with 0)
    IF NOT (NEW.address->>'pin' ~* '^[1-9][0-9]{5}$') THEN
        RAISE EXCEPTION 'Invalid PIN code format';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_kyc_address
    BEFORE INSERT OR UPDATE ON user_kyc
    FOR EACH ROW
    EXECUTE FUNCTION validate_address_jsonb();
