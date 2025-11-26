-- Remove seeded templates
DELETE FROM notification_templates WHERE name IN (
    'otp_sms',
    'transaction_alert_sms',
    'welcome_email',
    'transaction_alert_email',
    'kyc_update_email',
    'account_alert_email',
    'wallet_created_push',
    'transaction_alert_push',
    'security_alert_push',
    'welcome_inapp'
);
