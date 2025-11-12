#!/bin/bash

# Script to validate nginx configuration
# Usage: ./scripts/validate-config.sh

set -e

echo "Validating nginx configuration..."

# Check if nginx.conf exists
if [ ! -f "nginx.conf" ]; then
    echo "Error: nginx.conf not found"
    exit 1
fi

# Try to validate using Docker (if available)
if command -v docker &> /dev/null; then
    echo "Using Docker to validate configuration..."
    docker run --rm \
        -v "$(pwd)/nginx.conf:/etc/nginx/conf.d/default.conf:ro" \
        nginx:1.25-alpine \
        nginx -t
    
    if [ $? -eq 0 ]; then
        echo "✓ Nginx configuration is valid"
        exit 0
    else
        echo "✗ Nginx configuration has errors"
        exit 1
    fi
else
    echo "Docker not available. Skipping validation."
    echo "Configuration will be validated when the container starts."
    exit 0
fi

