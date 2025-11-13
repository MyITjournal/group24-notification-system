release: psql $DATABASE_URL -f database/migrations/add_last_notification_columns.sql && psql $DATABASE_URL -f database/migrations/create_simple_users_table.sql && node scripts/seed.js
web: npm run start:prod
