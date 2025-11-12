# API Gateway Quick Start

This guide will help you quickly set up and test the API Gateway with the orchestrator service.

## Prerequisites

- Docker and Docker Compose installed
- Ports 8080 (gateway), 8081 (orchestrator), 5432 (postgres), 6379 (redis), 9092 (kafka) available

## Quick Start

### 1. Start All Services

From the project root:

```bash
docker-compose -f docker-compose.gateway.yml up -d
```

This will start:
- API Gateway (port 8080)
- Orchestrator service (port 8081, internal 8080)
- PostgreSQL database
- Redis
- Kafka + Zookeeper
- Run database migrations automatically

### 2. Verify Services Are Running

```bash
# Check all containers
docker ps

# Check gateway logs
docker logs api-gateway

# Check orchestrator logs
docker logs orchestrator
```

### 3. Test the Gateway

#### Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

#### Create Notification (through gateway)
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Content-Type: application/json" \
  -H "X-Request-ID: test-123" \
  -d '{
    "id": "test-123",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {
      "user_name": "Test User",
      "app_name": "MyApp"
    }
  }'
```

Expected response (201 Created):
```json
{
  "notification_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "pending",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

#### Test 404 (Unknown Route)
```bash
curl http://localhost:8080/unknown
```

Expected response (404):
```json
{
  "error": {
    "success": false,
    "message": "Not Found",
    "error": "The requested resource was not found on this server"
  }
}
```

### 4. Test Direct Orchestrator Access (for comparison)

```bash
# Health check directly
curl http://localhost:8081/health

# Create notification directly
curl -X POST http://localhost:8081/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "id": "test-direct-123",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {
      "user_name": "Test User",
      "app_name": "MyApp"
    }
  }'
```

### 5. View Logs

```bash
# Gateway logs
docker logs -f api-gateway

# Orchestrator logs
docker logs -f orchestrator

# All services logs
docker-compose -f docker-compose.gateway.yml logs -f
```

### 6. Stop Services

```bash
docker-compose -f docker-compose.gateway.yml down

# To also remove volumes (clears database data)
docker-compose -f docker-compose.gateway.yml down -v
```

## Troubleshooting

### Gateway returns 502 Bad Gateway
- Check if orchestrator is running: `docker ps | grep orchestrator`
- Check orchestrator logs: `docker logs orchestrator`
- Verify orchestrator health: `curl http://localhost:8081/health`

### Gateway returns 504 Gateway Timeout
- Check orchestrator response time
- Increase timeout in nginx.conf if needed
- Check database and Kafka connectivity

### Services won't start
- Check port conflicts: `netstat -tulpn | grep -E '8080|8081|5432|6379|9092'`
- Check Docker logs: `docker-compose -f docker-compose.gateway.yml logs`
- Verify Docker network: `docker network ls`

### Database connection errors
- Wait for postgres to be fully ready (may take 10-15 seconds)
- Check postgres logs: `docker logs postgres`
- Verify migrations ran: `docker logs orchestrator-migrations`

## Next Steps

Once the gateway is working, proceed with:
1. Integration tests (Gateway → Orchestrator)
2. End-to-end tests (Full flow with Kafka)
3. Rate limiting tests
4. Error scenario tests

