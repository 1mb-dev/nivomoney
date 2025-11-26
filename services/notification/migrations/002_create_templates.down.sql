-- Remove foreign key constraint
ALTER TABLE notifications DROP CONSTRAINT IF EXISTS fk_notifications_template;

-- Drop notification_templates table
DROP TABLE IF EXISTS notification_templates;
