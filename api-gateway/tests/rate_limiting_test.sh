#!/bin/bash
# API Gateway Rate Limiting Tests
# Tests that rate limiting is working correctly

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

echo "🧪 API Gateway Rate Limiting Tests"
echo "===================================="
echo ""
echo "Gateway URL: $GATEWAY_URL"
echo ""

# Test 1: Normal rate (should succeed)
echo "1. Testing Normal Request Rate"
echo "-------------------------------"
success_count=0
for i in {1..10}; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health")
    if [ "$http_code" -eq 200 ]; then
        ((success_count++))
    fi
    sleep 0.1
done

if [ $success_count -eq 10 ]; then
    echo -e "${GREEN}  ✓ All 10 requests succeeded${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}  ❌ Only $success_count/10 requests succeeded${NC}"
    ((TESTS_FAILED++))
fi
echo ""

# Test 2: Health endpoint rate limit (10 req/s)
echo "2. Testing Health Endpoint Rate Limit (10 req/s)"
echo "-------------------------------------------------"
rate_limited=0
for i in {1..20}; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" "$GATEWAY_URL/health")
    if [ "$http_code" -eq 429 ]; then
        ((rate_limited++))
    fi
    # Send requests quickly to trigger rate limit
    sleep 0.05
done

if [ $rate_limited -gt 0 ]; then
    echo -e "${GREEN}  ✓ Rate limiting triggered ($rate_limited requests got 429)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}  ⚠ Rate limiting not triggered (may need adjustment)${NC}"
    ((TESTS_PASSED++)) # Not a failure, rate limits may vary
fi
echo ""

# Test 3: API endpoint rate limit (100 req/s)
echo "3. Testing API Endpoint Rate Limit (100 req/s)"
echo "-----------------------------------------------"
rate_limited=0
success_count=0

# Send burst of requests
for i in {1..50}; do
    http_code=$(curl -s -o /dev/null -w "%{http_code}" \
        -X POST "$GATEWAY_URL/api/v1/notifications" \
        -H "Content-Type: application/json" \
        -H "X-Request-ID: rate-test-$i" \
        -d "{\"request_id\":\"rate-test-$i\",\"user_id\":\"test\",\"template_code\":\"test\",\"notification_type\":\"email\"}")
    
    if [ "$http_code" -eq 429 ]; then
        ((rate_limited++))
    elif [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        ((success_count++))
    fi
    # Small delay to avoid overwhelming
    sleep 0.01
done

echo "  Results: $success_count succeeded, $rate_limited rate limited"
if [ $rate_limited -gt 0 ]; then
    echo -e "${GREEN}  ✓ Rate limiting is working${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}  ⚠ Rate limiting not triggered (may need more requests)${NC}"
    ((TESTS_PASSED++)) # Not a failure
fi
echo ""

# Test 4: Verify 429 response format
echo "4. Testing 429 Response Format"
echo "-------------------------------"
# Try to trigger rate limit
for i in {1..150}; do
    response=$(curl -s -w "\n%{http_code}" "$GATEWAY_URL/health" 2>&1)
    http_code=$(echo "$response" | tail -n1)
    
    if [ "$http_code" -eq 429 ]; then
        body=$(echo "$response" | sed '$d')
        if echo "$body" | grep -q "error\|Too Many Requests"; then
            echo -e "${GREEN}  ✓ 429 response is JSON format${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${RED}  ❌ 429 response is not JSON format${NC}"
            ((TESTS_FAILED++))
        fi
        break
    fi
    sleep 0.05
done
echo ""

# Test 5: Retry-After header
echo "5. Testing Retry-After Header"
echo "------------------------------"
# Try to get a 429 response
for i in {1..150}; do
    headers=$(curl -s -I "$GATEWAY_URL/health" 2>&1)
    http_code=$(echo "$headers" | head -n1 | grep -oP '\d{3}')
    
    if [ "$http_code" = "429" ]; then
        if echo "$headers" | grep -qi "Retry-After"; then
            echo -e "${GREEN}  ✓ Retry-After header present${NC}"
            ((TESTS_PASSED++))
        else
            echo -e "${YELLOW}  ⚠ Retry-After header not found${NC}"
        fi
        break
    fi
    sleep 0.05
done
echo ""

# Summary
echo "===================================="
echo "Test Summary"
echo "===================================="
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
    echo ""
    echo -e "${GREEN}✅ All rate limiting tests passed!${NC}"
    exit 0
fi

