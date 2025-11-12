# API Gateway

The API Gateway is implemented using Nginx and serves as the single entry point for all API requests. It routes requests to the appropriate backend services (orchestrator) and provides additional functionality like rate limiting, CORS, and security headers.

## Features

- **Reverse Proxy**: Routes requests to orchestrator service
- **Rate Limiting**: Prevents abuse with configurable rate limits
- **CORS Support**: Handles cross-origin requests
- **Security Headers**: Adds security headers to all responses
- **Request ID Propagation**: Forwards X-Request-ID header
- **Error Handling**: Custom error pages with JSON responses
- **Health Checks**: Fast health check routing
- **Load Balancing**: Ready for multiple orchestrator instances

## Configuration

### Environment Variables

The gateway can be configured via environment variables or by modifying `nginx.conf`:

- `GATEWAY_PORT`: Port for the gateway (default: 80)
- `ORCHESTRATOR_HOST`: Orchestrator service hostname (default: orchestrator)
- `ORCHESTRATOR_PORT`: Orchestrator service port (default: 8080)

### Rate Limits

- **API Endpoints**: 100 requests/second per IP (burst: 20)
- **Health Checks**: 10 requests/second per IP (burst: 5)

### Routes

- `/api/v1/*` → Orchestrator service
- `/health/*` → Orchestrator service (fast path)
- `/*` → 404 Not Found

## Running Locally

### Using Docker Compose

```bash
# From project root
docker-compose -f docker-compose.gateway.yml up -d
```

### Using Docker

```bash
# Build image
docker build -t api-gateway:latest ./api-gateway

# Run container
docker run -d \
  --name api-gateway \
  -p 8080:80 \
  --network notification-system-network \
  api-gateway:latest
```

## Testing

### Running Tests

The API Gateway includes a comprehensive test suite located in `tests/`:

#### Quick Test (Config Only)
```bash
cd api-gateway
make test-config
```

#### All Tests (Requires Running Services)
```bash
# Start services first
docker-compose -f docker-compose.gateway.yml up -d

# Wait for services to be ready (30-60 seconds)
sleep 30

# Run all tests
cd api-gateway
make test-all

# Or run tests individually:
make test-config      # Nginx config validation
make test-routing     # Routing and header tests
make test-rate-limit  # Rate limiting tests
make test-e2e         # End-to-end integration tests
```

#### Using Test Scripts Directly
```bash
cd api-gateway
./tests/nginx_config_test.sh
./tests/routing_test.sh
./tests/rate_limiting_test.sh
./tests/e2e/gateway_integration_test.sh

# Or run all at once:
./tests/run-all-tests.sh
```

### Manual Testing

#### Health Check

```bash
curl http://localhost:8080/health
```

#### Create Notification (through gateway)

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-123" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "request_id": "test-123",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {
      "name": "Test User"
    }
  }'
```

#### Test Rate Limiting

```bash
# Send multiple rapid requests
for i in {1..150}; do
  curl -X POST http://localhost:8080/api/v1/notifications \
    -H "Content-Type: application/json" \
    -d '{"request_id":"test-'$i'","user_id":"usr_123","template_code":"test","notification_type":"email"}'
done
# Should see 429 responses after rate limit
```

See `tests/README.md` for detailed test documentation.

## Configuration Files

- `nginx.conf`: Main nginx configuration
- `Dockerfile`: Docker image definition

## Network Requirements

The gateway needs to be on the same Docker network as the orchestrator service to communicate with it.

## Logs

Access logs: `/var/log/nginx/api-gateway-access.log` (JSON format)
Error logs: `/var/log/nginx/api-gateway-error.log`

## Production Considerations

For production, consider:

1. **SSL/TLS**: Add SSL certificates and configure HTTPS
2. **Authentication**: Add authentication middleware (JWT, API keys, etc.)
3. **Monitoring**: Integrate with monitoring tools
4. **Load Balancing**: Configure multiple orchestrator instances
5. **Caching**: Add response caching for appropriate endpoints
6. **IP Whitelisting**: Restrict access to specific IPs if needed
7. **Request/Response Transformation**: Add transformation rules if needed

