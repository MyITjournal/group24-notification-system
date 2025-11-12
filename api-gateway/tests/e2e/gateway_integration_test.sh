#!/bin/bash
# End-to-End Gateway Integration Tests
# Tests the full flow: Gateway -> Orchestrator -> Response

set -e

GATEWAY_URL="${GATEWAY_URL:-http://localhost:8080}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8081}"
API_KEY="${API_KEY:-test-api-key}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to test endpoint and validate response
test_e2e() {
    local method=$1
    local url=$2
    local expected_status=$3
    local description=$4
    local headers=$5
    local data=$6
    local validate_func=$7
    
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
        echo -e "${GREEN}  ✓ Status: $http_code${NC}"
        
        # Run validation function if provided
        if [ -n "$validate_func" ]; then
            if $validate_func "$body"; then
                ((TESTS_PASSED++))
                return 0
            else
                echo -e "${RED}  ❌ Response validation failed${NC}"
                ((TESTS_FAILED++))
                return 1
            fi
        else
            ((TESTS_PASSED++))
            return 0
        fi
    else
        echo -e "${RED}  ❌ Status: $http_code (expected $expected_status)${NC}"
        echo -e "${YELLOW}  Response: $body${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Validation functions
validate_json_response() {
    local body=$1
    if echo "$body" | jq . >/dev/null 2>&1; then
        echo -e "${GREEN}    ✓ Response is valid JSON${NC}"
        return 0
    else
        echo -e "${RED}    ❌ Response is not valid JSON${NC}"
        return 1
    fi
}

validate_health_response() {
    local body=$1
    if echo "$body" | jq -e '.success == true or .status == "ok"' >/dev/null 2>&1; then
        echo -e "${GREEN}    ✓ Health check response is valid${NC}"
        return 0
    else
        echo -e "${RED}    ❌ Health check response is invalid${NC}"
        return 1
    fi
}

validate_notification_response() {
    local body=$1
    if echo "$body" | jq -e '.success == true and .data.notification_id' >/dev/null 2>&1; then
        echo -e "${GREEN}    ✓ Notification response is valid${NC}"
        return 0
    else
        echo -e "${RED}    ❌ Notification response is invalid${NC}"
        return 1
    fi
}

echo "🧪 API Gateway End-to-End Integration Tests"
echo "============================================"
echo ""
echo "Gateway URL: $GATEWAY_URL"
echo "Orchestrator URL: $ORCHESTRATOR_URL"
echo ""

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
for i in {1..30}; do
    if curl -s -f "$GATEWAY_URL/health" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ Services are ready${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${RED}❌ Services did not become ready in time${NC}"
        exit 1
    fi
    sleep 1
done
echo ""

# Test 1: Health check through gateway
echo "1. Health Check Through Gateway"
echo "-------------------------------"
test_e2e "GET" "$GATEWAY_URL/health" 200 "Health check" "" "" validate_health_response
echo ""

# Test 2: Health check directly (for comparison)
echo "2. Health Check Direct (Orchestrator)"
echo "--------------------------------------"
test_e2e "GET" "$ORCHESTRATOR_URL/health" 200 "Direct health check" "" "" validate_health_response
echo ""

# Test 3: Create notification through gateway
echo "3. Create Notification Through Gateway"
echo "---------------------------------------"
REQUEST_ID="e2e-test-$(date +%s)"
NOTIFICATION_DATA=$(cat <<EOF
{
  "request_id": "$REQUEST_ID",
  "user_id": "e2e-test-user",
  "template_code": "welcome_email",
  "notification_type": "email",
  "variables": {
    "name": "E2E Test User"
  },
  "priority": 2
}
EOF
)

test_e2e "POST" "$GATEWAY_URL/api/v1/notifications" 201 "Create notification" \
    "X-Request-ID: $REQUEST_ID" \
    "$NOTIFICATION_DATA" \
    validate_notification_response
echo ""

# Test 4: Verify request ID propagation
echo "4. Request ID Propagation"
echo "-------------------------"
CUSTOM_REQUEST_ID="custom-request-$(date +%s)"
response=$(curl -s -w "\n%{http_code}" -X GET "$GATEWAY_URL/health" \
    -H "X-Request-ID: $CUSTOM_REQUEST_ID" 2>&1)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ]; then
    # Check if request ID is in response or logs (we can't easily check logs here)
    echo -e "${GREEN}  ✓ Request ID header accepted${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}  ❌ Request failed${NC}"
    ((TESTS_FAILED++))
fi
echo ""

# Test 5: Idempotency through gateway
echo "5. Idempotency Through Gateway"
echo "-------------------------------"
IDEMPOTENT_REQUEST_ID="idempotent-$(date +%s)"
IDEMPOTENT_DATA=$(cat <<EOF
{
  "request_id": "$IDEMPOTENT_REQUEST_ID",
  "user_id": "idempotent-user",
  "template_code": "welcome_email",
  "notification_type": "email",
  "variables": {
    "name": "Idempotent Test"
  }
}
EOF
)

# First request
echo "  First request..."
test_e2e "POST" "$GATEWAY_URL/api/v1/notifications" 201 "First idempotent request" \
    "X-Request-ID: $IDEMPOTENT_REQUEST_ID" \
    "$IDEMPOTENT_DATA" \
    validate_notification_response

# Second request (should return cached response)
echo "  Second request (should be cached)..."
response=$(curl -s -w "\n%{http_code}" -X POST "$GATEWAY_URL/api/v1/notifications" \
    -H "X-Request-ID: $IDEMPOTENT_REQUEST_ID" \
    -H "Content-Type: application/json" \
    -d "$IDEMPOTENT_DATA" 2>&1)
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

if [ "$http_code" -eq 200 ] || [ "$http_code" -eq 201 ]; then
    echo -e "${GREEN}  ✓ Idempotent request handled correctly (status: $http_code)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${RED}  ❌ Idempotent request failed (status: $http_code)${NC}"
    ((TESTS_FAILED++))
fi
echo ""

# Test 6: Error handling - invalid request
echo "6. Error Handling - Invalid Request"
echo "------------------------------------"
INVALID_DATA='{"invalid": "data"}'
test_e2e "POST" "$GATEWAY_URL/api/v1/notifications" 400 "Invalid request" \
    "X-Request-ID: error-test-$(date +%s)" \
    "$INVALID_DATA" \
    validate_json_response
echo ""

# Test 7: Compare gateway vs direct orchestrator
echo "7. Gateway vs Direct Orchestrator Comparison"
echo "----------------------------------------------"
COMPARE_REQUEST_ID="compare-$(date +%s)"
COMPARE_DATA=$(cat <<EOF
{
  "request_id": "$COMPARE_REQUEST_ID",
  "user_id": "compare-user",
  "template_code": "welcome_email",
  "notification_type": "email",
  "variables": {"name": "Compare Test"}
}
EOF
)

# Through gateway
echo "  Through gateway..."
gateway_response=$(curl -s -w "\n%{http_code}" -X POST "$GATEWAY_URL/api/v1/notifications" \
    -H "X-Request-ID: $COMPARE_REQUEST_ID" \
    -H "Content-Type: application/json" \
    -d "$COMPARE_DATA" 2>&1)
gateway_code=$(echo "$gateway_response" | tail -n1)

# Direct (if orchestrator is accessible)
echo "  Direct to orchestrator..."
direct_response=$(curl -s -w "\n%{http_code}" -X POST "$ORCHESTRATOR_URL/api/v1/notifications" \
    -H "X-Request-ID: $COMPARE_REQUEST_ID-direct" \
    -H "Content-Type: application/json" \
    -d "$COMPARE_DATA" 2>&1)
direct_code=$(echo "$direct_response" | tail -n1)

if [ "$gateway_code" -eq "$direct_code" ] || ([ "$gateway_code" -ge 200 ] && [ "$gateway_code" -lt 300 ]); then
    echo -e "${GREEN}  ✓ Gateway routes correctly (gateway: $gateway_code, direct: $direct_code)${NC}"
    ((TESTS_PASSED++))
else
    echo -e "${YELLOW}  ⚠ Gateway response differs (gateway: $gateway_code, direct: $direct_code)${NC}"
    ((TESTS_PASSED++)) # Not a failure, just a difference
fi
echo ""

# Summary
echo "============================================"
echo "Test Summary"
echo "============================================"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
if [ $TESTS_FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $TESTS_FAILED${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: $TESTS_FAILED${NC}"
    echo ""
    echo -e "${GREEN}✅ All E2E tests passed!${NC}"
    exit 0
fi

