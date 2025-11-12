-- Drop trigger
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_notifications_user_status;
DROP INDEX IF EXISTS idx_notifications_created_at;
DROP INDEX IF EXISTS idx_notifications_status;
DROP INDEX IF EXISTS idx_notifications_user_id;

-- Drop table
DROP TABLE IF EXISTS notifications;

-- Drop enums
DROP TYPE IF EXISTS notification_status_enum;
DROP TYPE IF EXISTS notification_type_enum;

