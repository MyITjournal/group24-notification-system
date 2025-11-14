# User Notification Preferences Microservice

A NestJS-based microservice for managing user notification preferences with JWT authentication, caching, pagination, and idempotency support.

## Features

- **JWT Authentication** - Secure token-based authentication with 24-hour expiration
- **Standardized Responses** - Consistent `{success, data, error, message, meta}` format
- **Pagination** - Query-based pagination for user lists (page/limit params)
- **Idempotency** - Request ID support to prevent duplicate operations
- **Redis Caching** - Automatic caching with 70-90% DB load reduction
- **Password Security** - Bcrypt hashing with 10 salt rounds
- **Batch Operations** - Retrieve up to 100 users with partial cache hits
- **Health Checks** - Kubernetes-ready liveness/readiness probes

## Quick Start

### 1. Install & Configure

```bash
npm install
```

Add `.env` file.

```bash
PORT=8081
DATABASE_URL=postgresql://postgres:password@localhost:5432/user_service
# REDIS_URL=redis://localhost:6379  # Optional, uses in-memory if not set
JWT_SECRET=your-super-secret-key-change-in-production
```

### 2. Start Server

```bash
npm run start:dev
```

Server runs on `http://localhost:8081`

## Architecture

```
user-service/
├── src/
│   ├── auth/                      # JWT authentication
│   │   ├── auth.controller.ts     # Login endpoint
│   │   ├── auth.service.ts        # Token generation & validation
│   │   ├── jwt.strategy.ts        # Passport JWT strategy
│   │   └── jwt-auth.guard.ts      # Route protection
│   │
│   ├── simple_users/              # User management
│   │   ├── simple_users.controller.ts  # REST endpoints
│   │   ├── simple_users.service.ts     # Business logic
│   │   ├── dto/simple_user.dto.ts      # DTOs & ApiResponse wrapper
│   │   └── entity/simple_user.entity.ts # Database model
│   │
│   ├── cache/                     # Redis caching
│   │   ├── cache_service.ts       # Cache + idempotency operations
│   │   └── cache_module.ts        # Redis/in-memory config
│   │
│   └── health/                    # Health checks
│       └── health_controller.ts   # /health endpoints
│
├── .env                           # Configuration
└── test/manual/                   # Test scripts
    └── test-auth-flow.js          # Auth flow test
```

## Testing

### Manual Scripts

```bash
# Test full auth flow
node test/manual/test-auth-flow.js

# Test all endpoints
node test/manual/test-both-modules.js

# Test specific features
node test/manual/test-cache.js
node test/manual/test-update-preferences.js
```

## Configuration (.env)

```bash
# Server
PORT=8081
NODE_ENV=development

# Database
DATABASE_URL=postgresql://user:pass@host:5432/db_name
# Or individual params: DB_HOST, DB_PORT, DB_USERNAME, DB_PASSWORD, DB_NAME

# Cache (optional - uses in-memory if not set)
REDIS_URL=redis://localhost:6379
CACHE_TTL=3600

# Authentication (REQUIRED)
JWT_SECRET=your-super-secret-key-min-32-chars
JWT_EXPIRES_IN=24h
```

## Key Features Explained

### Authentication

- All endpoints except registration & login require JWT
- Token expires in 24 hours
- Use `Authorization: Bearer <token>` header

### Pagination

```bash
GET /api/v1/users?page=2&limit=20
```

- Default: `page=1`, `limit=10`
- Max limit: 100
- Returns `meta` with pagination info

### Idempotency

```bash
POST /api/v1/users
X-Request-ID: unique-request-id-123
```

- Duplicate requests return cached response
- 24-hour cache TTL
- Prevents duplicate user creation

### Caching

- User preferences cached for 1 hour
- Automatic invalidation on updates
- Falls back to in-memory if Redis unavailable
- Batch operations support partial cache hits

## Database Schema

```sql
CREATE TABLE simple_users (
  user_id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  push_token VARCHAR(500),
  email_preference BOOLEAN DEFAULT TRUE,
  push_preference BOOLEAN DEFAULT TRUE,
  last_notification_email TIMESTAMP,
  last_notification_push TIMESTAMP,
  last_notification_id VARCHAR(100),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Technology Stack

- **NestJS** v11.0.1 - Framework
- **TypeORM** v0.3.27 - ORM
- **PostgreSQL** - Database
- **JWT** - Authentication (@nestjs/jwt, passport-jwt)
- **Redis** - Caching (optional, falls back to in-memory)
- **bcrypt** v6.0.0 - Password hashing
- **TypeScript** v5.7.3 - Language

## Development Commands

```bash
npm install           # Install dependencies
npm run start:dev     # Development mode (hot reload)
npm run build         # Build for production
npm run start:prod    # Run production build
npm test              # Run tests
npm run lint          # Lint code
```

## Troubleshooting

**Redis connection error?**

- Comment out `REDIS_URL` in `.env` to use in-memory cache
- Or start Redis: `redis-server`

**Unauthorized (401)?**

- Login in as a user in order to get an `access_token`. `POST /api/v1/auth/login`
- Fill the `access_token` into Authorization header: `Bearer <token>`

**Build errors**

- Delete `node_modules` and `dist` folders
- Run `npm install` and `npm run build`
