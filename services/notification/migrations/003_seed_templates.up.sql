-- Seed default notification templates for common use cases

-- OTP SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'otp_sms',
    'sms',
    '',
    'Your Nivo Money OTP is {{otp}}. Valid for {{validity_minutes}} minutes. Do not share this code. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- Transaction Alert SMS Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'transaction_alert_sms',
    'sms',
    '',
    'Txn Alert: Rs {{amount}} {{transaction_type}} on {{date}}. Wallet bal: Rs {{balance}}. Ref: {{transaction_id}}. - Nivo Money',
    1,
    NOW(),
    NOW()
);

-- Welcome Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'welcome_email',
    'email',
    'Welcome to Nivo Money, {{user_name}}!',
    'Dear {{user_name}},

Welcome to Nivo Money! We''re excited to have you on board.

Your account has been successfully created. You can now:
- Create wallets
- Send and receive money
- Track your transactions

Get started by creating your first wallet.

If you have any questions, our support team is here to help.

Best regards,
The Nivo Money Team',
    1,
    NOW(),
    NOW()
);

-- Transaction Alert Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'transaction_alert_email',
    'email',
    'Transaction Alert: {{transaction_type}} of ₹{{amount}}',
    'Dear {{user_name}},

This is to inform you about a recent transaction on your Nivo Money account:

Transaction Details:
- Type: {{transaction_type}}
- Amount: ₹{{amount}}
- Date: {{date}}
- Reference: {{transaction_id}}
- Description: {{description}}

Current Wallet Balance: ₹{{balance}}

If you did not authorize this transaction, please contact our support team immediately.

Best regards,
The Nivo Money Team',
    1,
    NOW(),
    NOW()
);

-- KYC Update Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'kyc_update_email',
    'email',
    'KYC Status Update - {{kyc_status}}',
    'Dear {{user_name}},

Your KYC (Know Your Customer) verification status has been updated:

Status: {{kyc_status}}

{{#if approved}}
Congratulations! Your account is now fully verified. You can now enjoy all features of Nivo Money without any limits.
{{else}}
{{#if rejected}}
Unfortunately, we were unable to verify your documents. Reason: {{rejection_reason}}

Please submit valid documents to complete your verification.
{{else}}
Your KYC verification is currently under review. We''ll notify you once the review is complete.
{{/if}}
{{/if}}

Thank you for choosing Nivo Money.

Best regards,
The Nivo Money Team',
    1,
    NOW(),
    NOW()
);

-- Account Alert Email Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'account_alert_email',
    'email',
    'Important Account Alert',
    'Dear {{user_name}},

This is an important alert regarding your Nivo Money account:

Alert Type: {{alert_type}}
Message: {{message}}
Date: {{date}}

If this activity was not authorized by you, please contact our support team immediately.

Best regards,
The Nivo Money Team',
    1,
    NOW(),
    NOW()
);

-- Wallet Created Push Notification Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'wallet_created_push',
    'push',
    'Wallet Created Successfully',
    'Your new {{currency}} wallet has been created. Start sending and receiving money now!',
    1,
    NOW(),
    NOW()
);

-- Transaction Alert Push Notification Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'transaction_alert_push',
    'push',
    'Transaction: ₹{{amount}}',
    '{{transaction_type}} of ₹{{amount}} completed. Balance: ₹{{balance}}',
    1,
    NOW(),
    NOW()
);

-- Security Alert Push Notification Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'security_alert_push',
    'push',
    'Security Alert',
    '{{message}}. If this wasn''t you, please secure your account immediately.',
    1,
    NOW(),
    NOW()
);

-- In-App Welcome Message Template
INSERT INTO notification_templates (id, name, channel, subject_template, body_template, version, created_at, updated_at)
VALUES (
    gen_random_uuid(),
    'welcome_inapp',
    'in_app',
    'Welcome to Nivo Money!',
    'Hi {{user_name}}! Welcome to Nivo Money. Create your first wallet to get started.',
    1,
    NOW(),
    NOW()
);
