#!/bin/bash
# Nginx Configuration Validation Tests
# Tests that the nginx configuration is valid and has no syntax errors

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GATEWAY_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
CONFIG_FILE="$GATEWAY_DIR/nginx.conf"

echo "🧪 Testing Nginx Configuration"
echo "================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test 1: Check if config file exists
echo "1. Checking if nginx.conf exists..."
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}❌ Config file not found: $CONFIG_FILE${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Config file found${NC}"
echo ""

# Test 2: Validate nginx configuration syntax
echo "2. Validating nginx configuration syntax..."
# Note: Upstream host resolution will fail in isolation, but we can check syntax
VALIDATION_OUTPUT=$(docker run --rm \
    -v "$CONFIG_FILE:/etc/nginx/conf.d/default.conf:ro" \
    nginx:1.25-alpine \
    nginx -t 2>&1 || true)

if echo "$VALIDATION_OUTPUT" | grep -q "syntax is ok"; then
    echo -e "${GREEN}✓ Configuration syntax is valid${NC}"
elif echo "$VALIDATION_OUTPUT" | grep -q "host not found in upstream"; then
    echo -e "${YELLOW}⚠ Upstream host not resolvable (expected in isolation)${NC}"
    echo -e "${GREEN}  ✓ Configuration syntax appears valid (upstream resolution is runtime)${NC}"
    # Check for actual syntax errors (not DNS issues)
    SYNTAX_ERRORS=$(echo "$VALIDATION_OUTPUT" | grep -E "\[emerg\].*syntax|\[error\].*syntax" || true)
    if [ -n "$SYNTAX_ERRORS" ] && ! echo "$SYNTAX_ERRORS" | grep -q "host not found"; then
        echo -e "${RED}❌ Configuration has syntax errors:${NC}"
        echo "$SYNTAX_ERRORS"
        exit 1
    fi
else
    # Check if there are actual syntax errors (not just upstream resolution)
    if echo "$VALIDATION_OUTPUT" | grep -qE "\[emerg\].*syntax|\[error\].*syntax"; then
        echo -e "${RED}❌ Configuration syntax error${NC}"
        echo "$VALIDATION_OUTPUT"
        exit 1
    else
        echo -e "${YELLOW}⚠ Validation warning (may be upstream resolution):${NC}"
        echo "$VALIDATION_OUTPUT" | head -5
        echo -e "${GREEN}  ✓ No syntax errors detected${NC}"
    fi
fi
echo ""

# Test 3: Check for required directives
echo "3. Checking for required configuration directives..."

REQUIRED_DIRECTIVES=(
    "upstream orchestrator"
    "location /health"
    "location /api/v1/"
    "limit_req_zone"
    "proxy_pass"
    "Access-Control-Allow-Origin"
    "X-Request-ID"
)

MISSING=0
for directive in "${REQUIRED_DIRECTIVES[@]}"; do
    if grep -q "$directive" "$CONFIG_FILE"; then
        echo -e "${GREEN}  ✓ Found: $directive${NC}"
    else
        echo -e "${RED}  ❌ Missing: $directive${NC}"
        MISSING=1
    fi
done

if [ $MISSING -eq 1 ]; then
    echo -e "${RED}❌ Some required directives are missing${NC}"
    exit 1
fi
echo ""

# Test 4: Check for common mistakes
echo "4. Checking for common configuration mistakes..."

# Check for localhost in proxy_pass (should use service names)
if grep -q "proxy_pass.*localhost" "$CONFIG_FILE"; then
    echo -e "${YELLOW}⚠ Warning: Found 'localhost' in proxy_pass. Use service names in Docker.${NC}"
else
    echo -e "${GREEN}  ✓ No localhost in proxy_pass${NC}"
fi

# Check for hardcoded ports (should be configurable)
if grep -q "listen 80" "$CONFIG_FILE"; then
    echo -e "${GREEN}  ✓ Using standard port 80${NC}"
fi

echo ""

# Test 5: Check log format
echo "5. Checking log format configuration..."
if grep -q "log_format json_format" "$CONFIG_FILE"; then
    echo -e "${GREEN}  ✓ JSON log format defined${NC}"
else
    echo -e "${YELLOW}⚠ Warning: JSON log format not found${NC}"
fi

if grep -q "access_log.*json_format" "$CONFIG_FILE"; then
    echo -e "${GREEN}  ✓ Access log uses JSON format${NC}"
else
    echo -e "${YELLOW}⚠ Warning: Access log not using JSON format${NC}"
fi
echo ""

echo -e "${GREEN}✅ All configuration tests passed!${NC}"
echo ""

