-- Create risk_rules table
CREATE TABLE IF NOT EXISTS risk_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    parameters JSONB NOT NULL,  -- Rule-specific parameters
    action VARCHAR(20) NOT NULL DEFAULT 'flag',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT risk_rules_type_check CHECK (rule_type IN ('velocity', 'daily_limit', 'threshold')),
    CONSTRAINT risk_rules_action_check CHECK (action IN ('allow', 'block', 'flag'))
);

-- Create indexes for risk_rules
CREATE INDEX idx_risk_rules_type ON risk_rules(rule_type);
CREATE INDEX idx_risk_rules_enabled ON risk_rules(enabled);
CREATE INDEX idx_risk_rules_created_at ON risk_rules(created_at DESC);

-- Create unique index for rule names
CREATE UNIQUE INDEX idx_risk_rules_name ON risk_rules(name);

-- Create trigger to update updated_at
CREATE TRIGGER update_risk_rules_updated_at
    BEFORE UPDATE ON risk_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create risk_events table for audit trail
CREATE TABLE IF NOT EXISTS risk_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL,  -- Reference to transaction being evaluated
    user_id UUID NOT NULL,          -- User being evaluated
    rule_id UUID,                   -- Rule that was triggered (NULL if no rules triggered)
    rule_type VARCHAR(50),          -- Type of rule triggered
    risk_score INTEGER NOT NULL DEFAULT 0,  -- Risk score (0-100)
    action VARCHAR(20) NOT NULL DEFAULT 'allow',
    reason TEXT NOT NULL,
    metadata JSONB,  -- Additional context
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT risk_events_score_check CHECK (risk_score >= 0 AND risk_score <= 100),
    CONSTRAINT risk_events_action_check CHECK (action IN ('allow', 'block', 'flag')),
    CONSTRAINT risk_events_type_check CHECK (
        (rule_id IS NOT NULL AND rule_type IS NOT NULL) OR
        (rule_id IS NULL AND rule_type IS NULL)
    ),
    FOREIGN KEY (rule_id) REFERENCES risk_rules(id) ON DELETE SET NULL
);

-- Create indexes for risk_events
CREATE INDEX idx_risk_events_transaction ON risk_events(transaction_id);
CREATE INDEX idx_risk_events_user ON risk_events(user_id);
CREATE INDEX idx_risk_events_rule ON risk_events(rule_id);
CREATE INDEX idx_risk_events_action ON risk_events(action);
CREATE INDEX idx_risk_events_created_at ON risk_events(created_at DESC);
CREATE INDEX idx_risk_events_risk_score ON risk_events(risk_score DESC);

-- Create index for high-risk events
CREATE INDEX idx_risk_events_high_risk
    ON risk_events(user_id, created_at DESC)
    WHERE risk_score >= 70;

-- Insert default risk rules
INSERT INTO risk_rules (name, rule_type, parameters, action, enabled) VALUES
    (
        'High Velocity Check',
        'velocity',
        '{"max_transactions": 5, "time_window_mins": 60, "per_user": true}'::jsonb,
        'flag',
        true
    ),
    (
        'Daily Limit - Individual',
        'daily_limit',
        '{"max_amount": 10000000, "currency": "INR", "per_user": true}'::jsonb,
        'block',
        true
    ),
    (
        'Large Transaction Alert',
        'threshold',
        '{"min_amount": 0, "max_amount": 5000000, "currency": "INR"}'::jsonb,
        'flag',
        true
    );
