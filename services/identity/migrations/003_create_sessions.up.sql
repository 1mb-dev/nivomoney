-- Create sessions table for active user sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of JWT token
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT sessions_expires_future CHECK (expires_at > created_at)
);

-- Create index on user_id for fast lookups
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- Create index on token_hash for validation
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);

-- Create index on expires_at for cleanup queries
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Function to clean up expired sessions (can be called periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM sessions WHERE expires_at < NOW();
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Create a scheduled cleanup job (requires pg_cron extension in production)
-- COMMENT: In production, use pg_cron to schedule: SELECT cleanup_expired_sessions();
