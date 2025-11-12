# Mock Services Testing Guide

Quick reference for testing with enhanced mocks.

## Quick Start

### Enable Realistic Behavior

```bash
# Enable delays and errors
export MOCK_SIMULATE_DELAY=true
export MOCK_SIMULATE_ERRORS=true
export MOCK_ERROR_RATE=0.2  # 20% error rate
```

### Test Scenarios

#### 1. Happy Path
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "happy-1",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "priority": 2,
    "variables": {
      "user_name": "John Doe",
      "app_name": "MyApp"
    }
  }'
```

#### 2. User Not Found
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "notfound-1",
    "user_id": "usr_notfound",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
```

#### 3. Email Disabled
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "noemail-1",
    "user_id": "usr_noemail_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
```

#### 4. Template Not Found
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "notemplate-1",
    "user_id": "usr_123",
    "template_code": "nonexistent_template",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
```

#### 5. Missing Required Variables
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "missingvars-1",
    "user_id": "usr_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test"}
    # Missing app_name
  }'
```

#### 6. Timeout Scenario
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "timeout-1",
    "user_id": "timeout_user_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
```

#### 7. Service Unavailable
```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "X-API-Key: test-key" \
  -d '{
    "request_id": "unavailable-1",
    "user_id": "unavailable_user_123",
    "template_code": "welcome_email",
    "notification_type": "email",
    "variables": {"user_name": "Test", "app_name": "MyApp"}
  }'
```

#### 8. Random Errors (with error rate)
```bash
# Set error rate
export MOCK_SIMULATE_ERRORS=true
export MOCK_ERROR_RATE=0.5  # 50% error rate

# Make multiple requests - some will randomly fail
for i in {1..20}; do
  curl -X POST http://localhost:8080/api/v1/notifications \
    -H "X-API-Key: test-key" \
    -d "{
      \"request_id\": \"random-$i\",
      \"user_id\": \"usr_$i\",
      \"template_code\": \"welcome_email\",
      \"notification_type\": \"email\",
      \"variables\": {\"user_name\": \"User $i\", \"app_name\": \"MyApp\"}
    }"
  echo ""
done
```

## User ID Patterns Reference

| Pattern | Result |
|---------|--------|
| `usr_123` | Normal user (both enabled) |
| `usr_noemail_*` | Email disabled |
| `usr_nopush_*` | Push disabled |
| `usr_nonotif_*` | Both disabled |
| `usr_emailonly_*` | Email only |
| `usr_pushonly_*` | Push only |
| `usr_notfound` | User not found |
| `error_*` | Random error |
| `timeout_*` | Timeout error |
| `unavailable_*` | Service unavailable |

## Template ID Patterns Reference

| Pattern | Result |
|---------|--------|
| `welcome_email` | ✅ Valid template |
| `password_reset` | ✅ Valid template |
| `push_notification` | ✅ Valid template |
| `order_confirmation` | ✅ Valid template |
| `nonexistent_template` | ❌ Not found |
| `notfound_*` | ❌ Not found |
| `timeout_*` | ⏱️ Timeout |
| `unavailable_*` | ❌ Service unavailable |
| `missing_vars_*` | ❌ Missing variables |
| `invalid_vars_*` | ❌ Invalid variables |
| `render_timeout_*` | ⏱️ Rendering timeout |

## Testing Checklist

- [ ] Happy path (normal user, valid template)
- [ ] User not found
- [ ] Template not found
- [ ] Email disabled user
- [ ] Push disabled user
- [ ] Both notifications disabled
- [ ] Missing required variables
- [ ] Timeout scenarios
- [ ] Service unavailable
- [ ] Random errors (with retry logic)
- [ ] Idempotency (same request_id twice)
- [ ] Different priority levels (0-4)

