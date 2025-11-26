-- Create notification_templates table
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    channel VARCHAR(20) NOT NULL,
    subject_template VARCHAR(500),
    body_template TEXT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT templates_channel_check CHECK (channel IN ('sms', 'email', 'push', 'in_app')),
    CONSTRAINT templates_version_check CHECK (version > 0)
);

-- Create indexes
CREATE INDEX idx_templates_name ON notification_templates(name);
CREATE INDEX idx_templates_channel ON notification_templates(channel);

-- Create trigger for updated_at
CREATE TRIGGER update_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add foreign key from notifications to templates
ALTER TABLE notifications
    ADD CONSTRAINT fk_notifications_template
    FOREIGN KEY (template_id)
    REFERENCES notification_templates(id)
    ON DELETE SET NULL;

-- Add comment
COMMENT ON TABLE notification_templates IS 'Reusable notification templates with variable placeholders';
COMMENT ON COLUMN notification_templates.name IS 'Unique template identifier (e.g., otp_sms, transaction_alert_email)';
COMMENT ON COLUMN notification_templates.version IS 'Template version for tracking changes';
