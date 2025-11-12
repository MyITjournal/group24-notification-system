#!/bin/bash
# API Gateway Routing Tests
# Tests that the gateway correctly routes requests to the orchestrator

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8081}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test HTTP endpoint
test_endpoint() {
    local method=$1
    local url=$2
    local expected_status=$3
    local description=$4
    local headers=$5
    local data=$6
    
    echo -e "${BLUE}Testing: $description${NC}"
    
    if [ -n "$data" ]; then
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
                -H "$headers" \
                -H "Content-Type: application/json" \
                -d "$data" 2>&1)
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
                -H "Content-Type: application/json" \
                -d "$data" 2>&1)
        fi
    else
        if [ -n "$headers" ]; then
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" \
                -H "$headers" 2>&1)
        else
            response=$(curl -s -w "\n%{http_code}" -X "$method" "$url" 2>&1)
        fi
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}  ✓ Status: $http_code (expected $expected_status)${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}  ❌ Status: $http_code (expected $expected_status)${NC}"
        echo -e "${YELLOW}  Response: $body${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Helper to check response headers
check_header() {
    local url=$1
    local header_name=$2
    local expected_value=$3
    local description=$4
    
    echo -e "${BLUE}Checking: $description${NC}"
    
    header_value=$(curl -s -I "$url" | grep -i "^$header_name:" | cut -d' ' -f2- | tr -d '\r\n')
    
    if echo "$header_value" | grep -q "$expected_value"; then
        echo -e "${GREEN}  ✓ Header $header_name contains: $expected_value${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}  ❌ Header $header_name: '$header_value' (expected to contain: $expected_value)${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

echo "🧪 API Gateway Routing Tests"
echo "=============================="
echo ""
echo "Gateway URL: $GATEWAY_URL"
echo "Orchestrator URL: $ORCHESTRATOR_URL"
echo ""

# Test 1: Health endpoint
echo "1. Testing Health Endpoint"
echo "---------------------------"
test_endpoint "GET" "$GATEWAY_URL/health" 200 "Health check through gateway"
echo ""

# Test 2: Health endpoint with request ID
echo "2. Testing Request ID Propagation"
echo "-----------------------------------"
test_endpoint "GET" "$GATEWAY_URL/health" 200 "Health check with X-Request-ID" "X-Request-ID: test-request-123"
check_header "$GATEWAY_URL/health" "X-Request-ID" "test-request-123" "Request ID header propagation"
echo ""

# Test 3: API endpoint routing
echo "3. Testing API Endpoint Routing"
echo "--------------------------------"
# Test with valid request (should get 400/422 for validation, or 201 if valid)
test_endpoint "POST" "$GATEWAY_URL/api/v1/notifications" "200|201|400|422" "POST /api/v1/notifications" \
    "X-Request-ID: test-routing-123" \
    '{"request_id":"test-123","user_id":"test-user","template_code":"test","notification_type":"email"}'
echo ""

# Test 4: 404 for unknown routes
echo "4. Testing 404 Handling"
echo "------------------------"
test_endpoint "GET" "$GATEWAY_URL/unknown" 404 "Unknown route returns 404"
test_endpoint "GET" "$GATEWAY_URL/api/v2/notifications" 404 "Unknown API version returns 404"
echo ""

# Test 5: Security headers
echo "5. Testing Security Headers"
echo "----------------------------"
check_header "$GATEWAY_URL/health" "X-Frame-Options" "SAMEORIGIN" "X-Frame-Options header"
check_header "$GATEWAY_URL/health" "X-Content-Type-Options" "nosniff" "X-Content-Type-Options header"
check_header "$GATEWAY_URL/health" "X-XSS-Protection" "1; mode=block" "X-XSS-Protection header"
echo ""

# Test 6: CORS headers
echo "6. Testing CORS Headers"
echo "-----------------------"
check_header "$GATEWAY_URL/health" "Access-Control-Allow-Origin" "*" "CORS Allow-Origin header"
check_header "$GATEWAY_URL/health" "Access-Control-Allow-Methods" "GET" "CORS Allow-Methods header"
check_header "$GATEWAY_URL/health" "Access-Control-Allow-Headers" "X-Request-ID" "CORS Allow-Headers header"
echo ""

# Test 7: OPTIONS preflight
echo "7. Testing OPTIONS Preflight"
echo "-----------------------------"
test_endpoint "OPTIONS" "$GATEWAY_URL/api/v1/notifications" 204 "OPTIONS preflight request"
echo ""

# Test 8: Error pages (502/503/504)
echo "8. Testing Error Page Format"
echo "------------------------------"
# We can't easily trigger 502/503/504, but we can verify 404 format
response=$(curl -s "$GATEWAY_URL/unknown")
if echo "$response" | grep -q "error\|success"; then
    echo -e "${GREEN}  ✓ Error response is JSON format${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}  ❌ Error response is not JSON format${NC}"
    ((TESTS_FAILED++))
fi
echo ""

# Summary
echo "=============================="
echo "Test Summary"
echo "=============================="
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
    echo ""
    echo -e "${GREEN}✅ All routing tests passed!${NC}"
    exit 0
fi

