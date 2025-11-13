-- Migration: Add last notification tracking columns
-- Date: 2025-11-11
-- Description: Adds three columns to the users table for tracking last notification times

-- Add the new columns to the users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS last_notification_email TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_notification_push TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_notification_id VARCHAR(100),
ADD COLUMN IF NOT EXISTS name VARCHAR(255) DEFAULT 'User',
ADD COLUMN IF NOT EXISTS password VARCHAR(255) DEFAULT '',
ADD COLUMN IF NOT EXISTS push_token TEXT,
ADD COLUMN IF NOT EXISTS email_preference BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS push_preference BOOLEAN DEFAULT true;

-- Add indexes for better query performance (optional but recommended)
CREATE INDEX IF NOT EXISTS idx_users_last_notification_email ON users(last_notification_email);
CREATE INDEX IF NOT EXISTS idx_users_last_notification_push ON users(last_notification_push);
CREATE INDEX IF NOT EXISTS idx_users_last_notification_id ON users(last_notification_id);

-- Verify the columns were added
SELECT column_name, data_type, is_nullable 
FROM information_schema.columns 
WHERE table_name = 'users' 
  AND column_name IN ('last_notification_email', 'last_notification_push', 'last_notification_id')
ORDER BY column_name;
