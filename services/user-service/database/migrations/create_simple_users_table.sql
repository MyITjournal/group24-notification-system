-- Migration: Create simple_users table
-- Purpose: Separate lightweight user table for simple authentication
-- Date: 2025-11-12

CREATE TABLE IF NOT EXISTS simple_users (
    user_id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    push_token TEXT,
    email_preference BOOLEAN DEFAULT true,
    push_preference BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_simple_users_email ON simple_users(email);

--: Add missing columns
ALTER TABLE simple_users 
ADD COLUMN IF NOT EXISTS last_notification_email TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_notification_push TIMESTAMP,
ADD COLUMN IF NOT EXISTS last_notification_id VARCHAR(100),
