#!/usr/bin/env node

/**
 * Database seeding script for Heroku deployment
 * This script inserts sample data into the database
 */

const { Client } = require('pg');

const seedData = async () => {
  const client = new Client({
    connectionString: process.env.DATABASE_URL,
    ssl:
      process.env.NODE_ENV === 'production'
        ? { rejectUnauthorized: false }
        : false,
  });

  try {
    await client.connect();
    console.log('Connected to database');

    // Check if sample data already exists
    const checkResult = await client.query(
      "SELECT COUNT(*) FROM users WHERE user_id = 'usr_7x9k2p'",
    );

    if (parseInt(checkResult.rows[0].count) > 0) {
      console.log('Sample data already exists. Skipping seed.');
      return;
    }

    console.log('Inserting sample data...');

    // Insert sample user
    await client.query(`
      INSERT INTO users (user_id, email, phone, timezone, language, notification_enabled, marketing, transactional, reminders, digest_enabled, digest_frequency, digest_time, last_notification_email, last_notification_push, last_notification_id)
      VALUES ('usr_7x9k2p', 'user@example.com', '+254712345678', 'Africa/Nairobi', 'en', true, false, true, true, true, 'daily', '09:00', '2025-01-15 09:30:00', '2025-01-15 10:15:00', 'notif_sample_123')
    `);

    // Insert email channel
    await client.query(`
      INSERT INTO user_channels (id, user_id, channel_type, enabled, verified, frequency, quiet_hours_enabled, quiet_hours_start, quiet_hours_end, quiet_hours_timezone)
      VALUES ('550e8400-e29b-41d4-a716-446655440001', 'usr_7x9k2p', 'email', true, true, 'immediate', true, '22:00', '07:00', 'Africa/Nairobi')
    `);

    // Insert push channel
    await client.query(`
      INSERT INTO user_channels (id, user_id, channel_type, enabled, verified, quiet_hours_enabled)
      VALUES ('550e8400-e29b-41d4-a716-446655440002', 'usr_7x9k2p', 'push', true, null, false)
    `);

    // Insert devices
    await client.query(`
      INSERT INTO user_devices (device_id, channel_id, platform, token, last_seen, active)
      VALUES 
        ('dev_abc123', '550e8400-e29b-41d4-a716-446655440002', 'ios', 'fcm_token_xyz...', '2025-01-15 10:25:00', true),
        ('dev_def456', '550e8400-e29b-41d4-a716-446655440002', 'android', 'fcm_token_abc...', '2025-01-14 18:30:00', true)
    `);

    console.log('Sample data inserted successfully!');
  } catch (error) {
    console.error('Error seeding database:', error.message);
    process.exit(1);
  } finally {
    await client.end();
  }
};

seedData();
