# User Service API Contract
## Integration with Orchestrator Service

This document specifies the API contract that the Orchestrator Service expects from the User Service for seamless integration.

---

## Endpoint: Get User Preferences

### Request

**Method:** `GET`  
**Path:** `/api/v1/users/{user_id}/preferences`  
**Content-Type:** `application/json`

**Path Parameters:**
- `user_id` (string, required): The unique identifier of the user

**Query Parameters:**
- Optional: `include_channels` (boolean, default: true) - Can be ignored by orchestrator, but should be supported for consistency

**Headers:**
- `Content-Type: application/json` (set by orchestrator)

**Request Body:** None

### Success Response

**Status Code:** `200 OK`

**Response Body Format:**
The response MUST be a direct JSON object (NOT wrapped in any envelope like `{ data: {...} }` or `{ success: true, ... }`).

{
  "email_enabled": true,
  "push_enabled": false
}**Field Specifications:**

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `email_enabled` | boolean | Yes | Whether email notifications are enabled for this user |
| `push_enabled` | boolean | Yes | Whether push notifications are enabled for this user |

**Important Notes:**
- Field names MUST use **snake_case** (`email_enabled`, `push_enabled`)
- The response MUST be a flat JSON object at the root level
- Additional fields are acceptable but will be ignored by the orchestrator
- Both fields MUST be present in the response

**Example Success Response:**son
{
  "email_enabled": true,
  "push_enabled": true
}---

## Error Responses

The orchestrator handles different error scenarios based on HTTP status codes. Error responses can be in any format, but the orchestrator will log the response body for debugging purposes.

### Non-Retryable Errors (4xx - Client Errors)

These errors indicate issues with the request that won't be resolved by retrying.

**Status Codes:**
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - User does not exist
- `422 Unprocessable Entity` - Validation errors

**Behavior:**
- Orchestrator will NOT retry these requests
- Error will be logged and returned to the caller
- Circuit breaker will NOT count these as failures

**Example 404 Response:**
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID abc123 does not exist",
    "details": {
      "user_id": "abc123"
    }
  }
}### Retryable Errors (5xx - Server Errors)

These errors indicate temporary server issues that may be resolved by retrying.

**Status Codes:**
- `500 Internal Server Error` - Server error
- `502 Bad Gateway` - Gateway error
- `503 Service Unavailable` - Service temporarily unavailable
- `504 Gateway Timeout` - Gateway timeout
- `429 Too Many Requests` - Rate limiting (also retryable)

**Behavior:**
- Orchestrator will automatically retry these requests with exponential backoff
- Retries will continue until max attempts are reached or request succeeds
- Circuit breaker will count these as failures
- After threshold failures, circuit breaker will open and stop sending requests

**Example 503 Response:**
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Unable to process request",
    "details": {
      "error": "Database connection timeout"
    }
  }
}---

## Timeout Requirements

- **Request Timeout:** The orchestrator uses a default timeout of 10 seconds
- **Response Time:** Responses should be returned within 2-3 seconds for optimal performance
- **Timeout Behavior:** If a request times out, it will be retried (treated as retryable error)

---

## Circuit Breaker Behavior

The orchestrator implements a circuit breaker pattern for resilience:

- **Circuit Open:** After multiple consecutive failures, the circuit opens and requests are immediately rejected
- **Circuit Half-Open:** After a timeout period, the circuit enters half-open state to test if service recovered
- **Circuit Closed:** Normal operation when service is healthy

**Recommendations:**
- Implement proper error handling to avoid cascading failures
- Return appropriate status codes (5xx for server errors, 4xx for client errors)
- Implement rate limiting with `429 Too Many Requests` instead of dropping requests

---

## Testing Checklist

Before integration, please verify:

- [ ] Success response returns `200 OK` with correct JSON structure
- [ ] Field names are exactly `email_enabled` and `push_enabled` (snake_case)
- [ ] Response is NOT wrapped in any envelope (no `data`, `success`, or `result` wrapper)
- [ ] Both boolean fields are always present in the response
- [ ] `404 Not Found` is returned for non-existent users
- [ ] `500` or `503` is returned for server errors (not `404` for server issues)
- [ ] Response time is under 3 seconds for normal requests
- [ ] Error responses include helpful error messages (for logging/debugging)

---

## Example Integration Test Cases

### Test Case 1: Valid User with Both Preferences Enabled
GET /api/v1/users/user-123/preferences
**Expected Response:**son
{
  "email_enabled": true,
  "push_enabled": true
}### Test Case 2: Valid User with Email Disabled
GET /api/v1/users/user-456/preferences**Expected Response:**on
{
  "email_enabled": false,
  "push_enabled": true
}
### Test Case 3: Non-existent User
GET /api/v1/users/non-existent-user/preferences**Expected Response:**tp
Status: 404 Not Found
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with ID non-existent-user does not exist",
    "details": {
      "user_id": "non-existent-user"
    }
  }
}### Test Case 4: Server Error
GET /api/v1/users/user-789/preferences**Expected Response:**
Status: 503 Service Unavailable
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Unable to process request",
    "details": {
      "error": "Database connection failed"
    }
  }
}
---

