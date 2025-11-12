#!/bin/bash
# Run all API Gateway tests
# This script runs all test suites in sequence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "🧪 API Gateway Test Suite"
echo "========================="
echo ""

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test and track results
run_test() {
    local test_name=$1
    local test_script=$2
    
    echo -e "${BLUE}Running: $test_name${NC}"
    echo "----------------------------------------"
    
    if [ -f "$test_script" ]; then
        if bash "$test_script"; then
            echo -e "${GREEN}✓ $test_name PASSED${NC}"
            ((PASSED_TESTS++))
        else
            echo -e "${RED}✗ $test_name FAILED${NC}"
            ((FAILED_TESTS++))
        fi
    else
        echo -e "${RED}✗ Test script not found: $test_script${NC}"
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
    echo ""
}

# Test 1: Configuration validation
run_test "Nginx Configuration Tests" "tests/nginx_config_test.sh"

# Check if services are running
echo -e "${YELLOW}Checking if services are running...${NC}"
if curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "${GREEN}✓ Services are running${NC}"
    echo ""
    
    # Test 2: Routing tests
    run_test "Routing Tests" "tests/routing_test.sh"
    
    # Test 3: Rate limiting tests
    run_test "Rate Limiting Tests" "tests/rate_limiting_test.sh"
    
    # Test 4: E2E tests
    run_test "End-to-End Integration Tests" "tests/e2e/gateway_integration_test.sh"
else
    echo -e "${YELLOW}⚠ Services are not running${NC}"
    echo "Skipping integration tests."
    echo ""
    echo "To run integration tests, start services with:"
    echo "  docker-compose -f docker-compose.gateway.yml up -d"
    echo "  # Wait for services to be ready, then run:"
    echo "  ./tests/routing_test.sh"
    echo "  ./tests/rate_limiting_test.sh"
    echo "  ./tests/e2e/gateway_integration_test.sh"
    echo ""
fi

# Summary
echo "========================="
echo "Test Summary"
echo "========================="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    exit 1
else
    echo -e "${GREEN}Failed: $FAILED_TESTS${NC}"
    echo ""
    echo -e "${GREEN}✅ All tests passed!${NC}"
    exit 0
fi

