-- Create the database
CREATE DATABASE user_service;

-- Connect to the database
\c user_service;

-- Create users table
CREATE TABLE users (
    user_id VARCHAR(50) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(20),
    timezone VARCHAR(50) DEFAULT 'UTC' NOT NULL,
    language VARCHAR(10) DEFAULT 'en' NOT NULL,
    notification_enabled BOOLEAN DEFAULT true NOT NULL,
    marketing BOOLEAN DEFAULT false NOT NULL,
    transactional BOOLEAN DEFAULT true NOT NULL,
    reminders BOOLEAN DEFAULT true NOT NULL,
    digest_enabled BOOLEAN DEFAULT false NOT NULL,
    digest_frequency VARCHAR(20) DEFAULT 'daily' NOT NULL,
    digest_time VARCHAR(5) DEFAULT '09:00' NOT NULL,
    last_notification_email TIMESTAMP,
    last_notification_push TIMESTAMP,
    last_notification_id VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create user_channels table
CREATE TABLE user_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(50) NOT NULL,
    channel_type VARCHAR(20) NOT NULL,
    enabled BOOLEAN DEFAULT true NOT NULL,
    verified BOOLEAN DEFAULT false,
    frequency VARCHAR(20),
    quiet_hours_enabled BOOLEAN DEFAULT false,
    quiet_hours_start VARCHAR(5),
    quiet_hours_end VARCHAR(5),
    quiet_hours_timezone VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create user_devices table
CREATE TABLE user_devices (
    device_id VARCHAR(50) PRIMARY KEY,
    channel_id UUID NOT NULL,
    platform VARCHAR(20) NOT NULL,
    token TEXT NOT NULL,
    last_seen TIMESTAMP,
    active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (channel_id) REFERENCES user_channels(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_user_channels_user_id ON user_channels(user_id);
CREATE INDEX idx_user_channels_type ON user_channels(channel_type);
CREATE INDEX idx_user_devices_channel_id ON user_devices(channel_id);
CREATE INDEX idx_user_devices_active ON user_devices(active);

-- Insert sample data
INSERT INTO users (user_id, email, phone, timezone, language, notification_enabled, marketing, transactional, reminders, digest_enabled, digest_frequency, digest_time, last_notification_email, last_notification_push, last_notification_id)
VALUES ('usr_7x9k2p', 'user@example.com', '+254712345678', 'Africa/Nairobi', 'en', true, false, true, true, true, 'daily', '09:00', '2025-01-15 09:30:00', '2025-01-15 10:15:00', 'notif_sample_123');

-- Insert email channel
INSERT INTO user_channels (id, user_id, channel_type, enabled, verified, frequency, quiet_hours_enabled, quiet_hours_start, quiet_hours_end, quiet_hours_timezone)
VALUES ('550e8400-e29b-41d4-a716-446655440001', 'usr_7x9k2p', 'email', true, true, 'immediate', true, '22:00', '07:00', 'Africa/Nairobi');

-- Insert push channel
INSERT INTO user_channels (id, user_id, channel_type, enabled, verified, quiet_hours_enabled)
VALUES ('550e8400-e29b-41d4-a716-446655440002', 'usr_7x9k2p', 'push', true, null, false);

-- Insert devices for push channel
INSERT INTO user_devices (device_id, channel_id, platform, token, last_seen, active)
VALUES 
    ('dev_abc123', '550e8400-e29b-41d4-a716-446655440002', 'ios', 'fcm_token_xyz...', '2025-01-15 10:25:00', true),
    ('dev_def456', '550e8400-e29b-41d4-a716-446655440002', 'android', 'fcm_token_abc...', '2025-01-14 18:30:00', true);
