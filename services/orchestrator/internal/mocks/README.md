# Mock Services Documentation

Enhanced mock services that simulate realistic production behavior including errors, delays, and edge cases.

## Features

- **Realistic Network Delays**: Simulates network latency (50-200ms by default)
- **Error Simulation**: Configurable error rates and specific error scenarios
- **Edge Cases**: Handles missing data, invalid inputs, timeouts, etc.
- **Configurable**: Behavior controlled via environment variables
- **Testing Support**: Special user/template IDs trigger specific behaviors

## Configuration

### Environment Variables

```bash
# Enable delay simulation (default: false)
MOCK_SIMULATE_DELAY=true

# Enable error simulation (default: false)
MOCK_SIMULATE_ERRORS=true

# Error rate (0.0 to 1.0, default: 0.1 = 10%)
MOCK_ERROR_RATE=0.15

# Delay range (default: 50ms-200ms)
MOCK_DELAY_MIN=50ms
MOCK_DELAY_MAX=200ms
```

## User Service Mock

### Normal Behavior

Returns user preferences based on user ID patterns:

```go
// Default: Both email and push enabled
userID := "usr_123"
prefs, err := mock.GetPreferences(userID)
// Returns: Email=true, Push=true
```

### Special User IDs (Error Scenarios)

| User ID Pattern | Behavior |
|----------------|----------|
| `error_*` | Random error (500, 503, timeout, etc.) |
| `notfound_*` or `usr_notfound` | User not found (404) |
| `timeout_*` | Request timeout (simulates slow service) |
| `unavailable_*` | Service unavailable (503) |
| `invalid_*` | Invalid response format |
| `no_email_*` or `usr_noemail_*` | Email disabled, push enabled |
| `no_push_*` or `usr_nopush_*` | Push disabled, email enabled |
| `no_notifications_*` or `usr_nonotif_*` | Both disabled |
| `email_only_*` or `usr_emailonly_*` | Email only |
| `push_only_*` or `usr_pushonly_*` | Push only |

### Examples

```bash
# User with email disabled
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-1",
    "user_id": "usr_noemail_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
  }'
# Should fail: email notifications disabled

# User not found
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-2",
    "user_id": "usr_notfound",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
  }'
# Should fail: user not found

# Simulate timeout
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-3",
    "user_id": "timeout_user_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
  }'
# Should trigger retry logic
```

## Template Service Mock

### Available Templates

- `welcome_email` - Welcome email template
- `password_reset` - Password reset email
- `push_notification` - Push notification template
- `order_confirmation` - Order confirmation email

### Special Template IDs (Error Scenarios)

| Template ID Pattern | Behavior |
|---------------------|----------|
| `notfound_*` or `nonexistent_template` | Template not found (404) |
| `timeout_*` | Request timeout |
| `unavailable_*` | Service unavailable (503) |
| `invalid_*` | Invalid template format |
| `*_v999` | Version not found |
| `missing_vars_*` | Missing required variables error |
| `invalid_vars_*` | Invalid variable types error |
| `render_timeout_*` | Rendering timeout |

### Examples

```bash
# Template not found
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-4",
    "user_id": "usr_123",
    "template_code": "notfound_template",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
  }'
# Should fail: template not found

# Missing required variables
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-5",
    "user_id": "usr_123",
    "template_code": "missing_vars_welcome_email",
    "notification_type": "email",
    "variables": {}  # Missing required variables
  }'
# Should fail: missing required variables

# Simulate timeout
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-6",
    "user_id": "usr_123",
    "template_code": "timeout_welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
# Should trigger retry logic
```

## Testing Scenarios

### 1. Test Retry Logic

```bash
# Enable error simulation
export MOCK_SIMULATE_ERRORS=true
export MOCK_ERROR_RATE=0.5  # 50% error rate

# Make multiple requests - some will fail and retry
for i in {1..10}; do
  curl -X POST http://localhost:8080/api/v1/notifications \
    -H "X-API-Key: test-key" \
    -H "Content-Type: application/json" \
    -d "{
      \"request_id\": \"test-retry-$i\",
      \"user_id\": \"usr_$i\",
      \"template_code\": \"welcome_email\",
      \"notification_type\": \"email\",
      \"variables\": {\"user_name\": \"User $i\", \"app_name\": \"MyApp\"}
    }"
done
```

### 2. Test Circuit Breaker

```bash
# Use failure threshold pattern
# (Would need to modify mock to support this via config)
```

### 3. Test User Preferences

```bash
# Test email disabled
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-email-disabled",
    "user_id": "usr_noemail_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
# Expected: Should fail with "email notifications disabled"

# Test push disabled
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-push-disabled",
    "user_id": "usr_nopush_123",
    "template_code": "push_notification",
    "notification_type": "push",
    "variables": {"title": "Test", "message": "Hello"}
  }'
# Expected: Should fail with "push notifications disabled"
```

### 4. Test Missing Variables

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-missing-vars",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
    # Missing app_name (required)
  }'
# Expected: Should fail with "missing required variable: app_name"
```

### 5. Test Timeout Scenarios

```bash
# User service timeout
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-user-timeout",
    "user_id": "timeout_user_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
# Expected: Should trigger retry, then fail after max retries

# Template service timeout
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "test-template-timeout",
    "user_id": "usr_123",
    "template_code": "timeout_welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
# Expected: Should trigger retry, then fail after max retries
```

## Error Types Simulated

### User Service Errors
- Internal server error (500)
- Service unavailable (503)
- Request timeout
- Database error
- Network error (connection refused)
- User not found (404)

### Template Service Errors
- Internal server error (500)
- Service unavailable (503)
- Request timeout
- Database error
- Network error
- Template not found (404)
- Version not found
- Missing required variables
- Invalid variable types
- Rendering timeout
- Template data corrupted

## Production-Like Behavior

1. **Network Latency**: 50-200ms delays simulate real network conditions
2. **Random Errors**: Configurable error rate simulates transient failures
3. **Timeout Scenarios**: Long delays test retry and circuit breaker logic
4. **Edge Cases**: Missing data, invalid inputs, service unavailability
5. **Realistic Responses**: Proper data structures matching production APIs

## Usage in Tests

```go
// Create mock with custom config
mock := mocks.NewUserServiceMockWithConfig(mocks.MockBehaviorConfig{
    SimulateDelay:    true,
    SimulateErrors:   true,
    ErrorRate:        0.2,
    DelayMin:         100 * time.Millisecond,
    DelayMax:         300 * time.Millisecond,
    FailureThreshold: 10, // Fail after 10 requests
})

// Use in tests
prefs, err := mock.GetPreferences("usr_123")
```

## Notes

- Mocks are thread-safe for concurrent testing
- Request counts are tracked for testing purposes
- Reset() method clears state between tests
- Special IDs are case-sensitive
- Error simulation is random but can be controlled via error rate

