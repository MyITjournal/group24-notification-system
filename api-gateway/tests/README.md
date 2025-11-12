# API Gateway Tests

This directory contains comprehensive tests for the API Gateway.

## Test Files

### 1. `nginx_config_test.sh`
Validates the Nginx configuration file:
- Checks if config file exists
- Validates syntax using Docker
- Verifies required directives are present
- Checks for common configuration mistakes
- Validates log format configuration

**Usage:**
```bash
./tests/nginx_config_test.sh
```

### 2. `routing_test.sh`
Tests that the gateway correctly routes requests:
- Health endpoint routing
- Request ID propagation
- API endpoint routing (`/api/v1/*`)
- 404 handling for unknown routes
- Security headers (X-Frame-Options, X-Content-Type-Options, etc.)
- CORS headers
- OPTIONS preflight requests
- Error page format (JSON)

**Usage:**
```bash
# Ensure gateway is running
docker-compose -f docker-compose.gateway.yml up -d

# Run tests
./tests/routing_test.sh
```

**Environment Variables:**
- `GATEWAY_URL`: Gateway URL (default: `http://localhost:8080`)
- `ORCHESTRATOR_URL`: Direct orchestrator URL (default: `http://localhost:8081`)

### 3. `rate_limiting_test.sh`
Tests rate limiting functionality:
- Normal request rate (should succeed)
- Health endpoint rate limit (10 req/s)
- API endpoint rate limit (100 req/s)
- 429 response format validation
- Retry-After header presence

**Usage:**
```bash
# Ensure gateway is running
docker-compose -f docker-compose.gateway.yml up -d

# Run tests
./tests/rate_limiting_test.sh
```

**Note:** Rate limiting tests may need adjustment based on actual rate limit configuration.

### 4. `e2e/gateway_integration_test.sh`
End-to-end integration tests:
- Full flow: Gateway → Orchestrator → Response
- Health check through gateway vs direct
- Create notification through gateway
- Request ID propagation
- Idempotency through gateway
- Error handling
- Gateway vs direct orchestrator comparison

**Usage:**
```bash
# Ensure full stack is running
docker-compose -f docker-compose.gateway.yml up -d

# Wait for services to be ready
sleep 30

# Run E2E tests
./tests/e2e/gateway_integration_test.sh
```

**Environment Variables:**
- `GATEWAY_URL`: Gateway URL (default: `http://localhost:8080`)
- `ORCHESTRATOR_URL`: Direct orchestrator URL (default: `http://localhost:8081`)
- `API_KEY`: API key for authentication (default: `test-api-key`)

## Running All Tests

### Option 1: Run individually
```bash
cd api-gateway
./tests/nginx_config_test.sh
./tests/routing_test.sh
./tests/rate_limiting_test.sh
./tests/e2e/gateway_integration_test.sh
```

### Option 2: Run via Makefile (if added)
```bash
cd api-gateway
make test-all
```

## Prerequisites

1. **Docker and Docker Compose** must be installed
2. **curl** must be installed
3. **jq** (optional but recommended) for JSON validation
4. **Services must be running** for integration/E2E tests:
   ```bash
   docker-compose -f docker-compose.gateway.yml up -d
   ```

## Test Requirements

### For Config Tests
- Nginx Docker image available
- Config file exists at `api-gateway/nginx.conf`

### For Routing/Rate Limiting/E2E Tests
- Gateway service running on port 8080
- Orchestrator service running and accessible
- Full stack (PostgreSQL, Redis, Kafka) running

## Troubleshooting

### Tests fail with "connection refused"
- Ensure services are running: `docker-compose -f docker-compose.gateway.yml ps`
- Check gateway logs: `docker logs api-gateway`
- Verify ports are not in use: `netstat -tuln | grep 8080`

### Rate limiting tests don't trigger
- Rate limits may be too high for the test
- Adjust the number of requests or timing in the test script
- Check nginx configuration for actual rate limit values

### JSON validation fails
- Install `jq`: `sudo apt-get install jq` (Ubuntu/Debian) or `brew install jq` (macOS)
- Or modify tests to use `grep` instead of `jq`

## Expected Results

- **Config tests**: Should always pass if config is valid
- **Routing tests**: Should pass when gateway and orchestrator are running
- **Rate limiting tests**: May show warnings if rate limits are not triggered (not a failure)
- **E2E tests**: Should pass when full stack is running and healthy

