-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    channel VARCHAR(20) NOT NULL,
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(500),
    body TEXT NOT NULL,
    template_id UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    correlation_id VARCHAR(100),
    source_service VARCHAR(50) NOT NULL,
    metadata JSONB,
    retry_count INT NOT NULL DEFAULT 0,
    failure_reason TEXT,
    queued_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT notifications_channel_check CHECK (channel IN ('sms', 'email', 'push', 'in_app')),
    CONSTRAINT notifications_priority_check CHECK (priority IN ('critical', 'high', 'normal', 'low')),
    CONSTRAINT notifications_status_check CHECK (status IN ('queued', 'sent', 'delivered', 'failed')),
    CONSTRAINT notifications_retry_count_check CHECK (retry_count >= 0 AND retry_count <= 10)
);

-- Create indexes for efficient queries
CREATE INDEX idx_notifications_user_id ON notifications(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_notifications_channel ON notifications(channel);
CREATE INDEX idx_notifications_type ON notifications(type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_correlation_id ON notifications(correlation_id) WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_notifications_source_service ON notifications(source_service);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_priority_status ON notifications(priority, status) WHERE status = 'queued'; -- For worker queue

-- Create unique constraint on correlation_id for idempotency
CREATE UNIQUE INDEX idx_notifications_correlation_id_unique ON notifications(correlation_id) WHERE correlation_id IS NOT NULL;

-- Create trigger to update updated_at automatically
CREATE TRIGGER update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comment
COMMENT ON TABLE notifications IS 'Stores all outbound notifications for simulation and audit';
COMMENT ON COLUMN notifications.correlation_id IS 'For idempotency - prevents duplicate notifications';
COMMENT ON COLUMN notifications.priority IS 'Critical (OTP) processed first, then high, normal, low';
COMMENT ON COLUMN notifications.retry_count IS 'Number of retry attempts (max 10)';
