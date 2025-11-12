.PHONY: help build up down logs test validate

# Variables
GATEWAY_IMAGE=api-gateway:latest
GATEWAY_PORT=8080

help:
	@echo "API Gateway - Available Commands:"
	@echo ""
	@echo "  make build      - Build the Docker image"
	@echo "  make up          - Start the gateway and dependencies"
	@echo "  make down        - Stop the gateway and dependencies"
	@echo "  make logs        - View gateway logs"
	@echo "  make validate    - Validate nginx configuration"
	@echo "  make test        - Run basic connectivity tests"

build:
	@echo "🐳 Building API Gateway Docker image..."
	@docker build -t $(GATEWAY_IMAGE) .

up:
	@echo "🚀 Starting API Gateway and dependencies..."
	@cd .. && docker-compose -f docker-compose.gateway.yml up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "✓ Services should be ready. Gateway available at http://localhost:$(GATEWAY_PORT)"

down:
	@echo "⏹️  Stopping API Gateway and dependencies..."
	@cd .. && docker-compose -f docker-compose.gateway.yml down

logs:
	@echo "📋 Showing API Gateway logs..."
	@docker logs -f api-gateway

validate:
	@echo "✅ Validating nginx configuration..."
	@docker run --rm -v $$(pwd)/nginx.conf:/etc/nginx/conf.d/default.conf:ro nginx:1.25-alpine nginx -t

test:
	@echo "🧪 Testing API Gateway connectivity..."
	@echo ""
	@echo "1. Testing health endpoint..."
	@curl -s http://localhost:$(GATEWAY_PORT)/health | head -20 || echo "❌ Health check failed"
	@echo ""
	@echo "2. Testing 404 endpoint..."
	@curl -s http://localhost:$(GATEWAY_PORT)/unknown | head -20 || echo "❌ 404 test failed"
	@echo ""
	@echo "✓ Basic connectivity tests complete"

clean:
	@echo "🧹 Cleaning up..."
	@docker rmi $(GATEWAY_IMAGE) 2>/dev/null || true

