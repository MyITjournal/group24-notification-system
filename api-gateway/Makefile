.PHONY: help build up down logs test validate test-config test-routing test-rate-limit test-e2e test-all

# Variables
GATEWAY_IMAGE=api-gateway:latest
GATEWAY_PORT=8080

help:
	@echo "API Gateway - Available Commands:"
	@echo ""
	@echo "  make build           - Build the Docker image"
	@echo "  make up               - Start the gateway and dependencies"
	@echo "  make down             - Stop the gateway and dependencies"
	@echo "  make logs             - View gateway logs"
	@echo "  make validate         - Validate nginx configuration"
	@echo "  make test             - Run basic connectivity tests"
	@echo "  make test-config      - Run nginx config validation tests"
	@echo "  make test-routing     - Run routing tests (requires running services)"
	@echo "  make test-rate-limit  - Run rate limiting tests (requires running services)"
	@echo "  make test-e2e         - Run end-to-end integration tests (requires running services)"
	@echo "  make test-all         - Run all tests (requires running services)"

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

test-config:
	@echo "🧪 Running nginx configuration tests..."
	@./tests/nginx_config_test.sh

test-routing:
	@echo "🧪 Running routing tests..."
	@./tests/routing_test.sh

test-rate-limit:
	@echo "🧪 Running rate limiting tests..."
	@./tests/rate_limiting_test.sh

test-e2e:
	@echo "🧪 Running end-to-end integration tests..."
	@./tests/e2e/gateway_integration_test.sh

test-all: test-config
	@echo ""
	@echo "⚠️  Note: The following tests require services to be running."
	@echo "   Start services with: make up"
	@echo "   Or: docker-compose -f ../docker-compose.gateway.yml up -d"
	@echo ""
	@read -p "Are services running? (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo ""; \
		$(MAKE) test-routing; \
		$(MAKE) test-rate-limit; \
		$(MAKE) test-e2e; \
	else \
		echo "Skipping integration tests. Start services and run them individually."; \
	fi

clean:
	@echo "🧹 Cleaning up..."
	@docker rmi $(GATEWAY_IMAGE) 2>/dev/null || true

