Reviewing the orchestrator's template service expectations to draft the API contract:


[3 tools called]


Reading the template model file for complete structure:


[2 tools called]


Template Service API contract:

```markdown
# Template Service API Contract
## Integration with Orchestrator Service

This document specifies the API contract that the Orchestrator Service expects from the Template Service for seamless integration.

---

## Endpoint 1: Get Template

### Request

**Method:** `GET`  
**Path:** `/api/v1/templates/{template_id}`  
**Content-Type:** `application/json`

**Path Parameters:**
- `template_id` (string, required): The unique identifier of the template

**Query Parameters:**
- `language` (string, required): The language code for the template (e.g., "en", "es", "fr")

**Headers:**
- `Content-Type: application/json` (set by orchestrator)

**Request Body:** None

**Example Request:**
```
GET /api/v1/templates/welcome_email?language=en
```

### Success Response

**Status Code:** `200 OK`

**Response Body Format:**
The response MUST be a direct JSON object (NOT wrapped in any envelope like `{ data: {...} }` or `{ success: true, ... }`).

```json
{
  "template_id": "welcome_email",
  "name": "Welcome Email",
  "version": "1.0.0",
  "language": "en",
  "type": "email",
  "subject": "Welcome {{name}}!",
  "body": {
    "html": "<h1>Welcome {{name}}!</h1><p>Click here: {{link}}</p>",
    "text": "Welcome {{name}}! Click here: {{link}}"
  },
  "variables": [
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "User's name"
    },
    {
      "name": "link",
      "type": "string",
      "required": true,
      "description": "Verification link"
    }
  ],
  "metadata": {
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-15T00:00:00Z",
    "created_by": "admin",
    "tags": ["welcome", "onboarding"]
  }
}
```

**Field Specifications:**

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `template_id` | string | Yes | Unique identifier of the template |
| `name` | string | Yes | Display name of the template |
| `version` | string | Yes | Version of the template |
| `language` | string | Yes | Language code (e.g., "en", "es") |
| `type` | string | Yes | Template type (e.g., "email", "push") |
| `subject` | string | No | Email subject line (for email templates) |
| `body` | object | Yes | Template body content |
| `body.html` | string | No | HTML version of the body |
| `body.text` | string | Yes | Plain text version of the body |
| `variables` | array | Yes | List of template variables |
| `variables[].name` | string | Yes | Variable name (used in template as `{{name}}`) |
| `variables[].type` | string | Yes | Variable type (e.g., "string", "number") |
| `variables[].required` | boolean | Yes | Whether the variable is required |
| `variables[].description` | string | Yes | Description of the variable |
| `metadata` | object | Yes | Template metadata |
| `metadata.created_at` | string (ISO 8601) | Yes | Creation timestamp |
| `metadata.updated_at` | string (ISO 8601) | Yes | Last update timestamp |
| `metadata.created_by` | string | Yes | Creator identifier |
| `metadata.tags` | array of strings | Yes | Template tags |

**Important Notes:**
- Field names MUST use **snake_case** (e.g., `template_id`, `created_at`)
- The response MUST be a flat JSON object at the root level (no wrapper)
- `subject` is optional and only present for email templates
- `body.html` is optional, but `body.text` is required
- Timestamps MUST be in ISO 8601 format (e.g., "2025-01-01T00:00:00Z")
- Variable placeholders in templates use double curly braces: `{{variable_name}}`

---

## Endpoint 2: Render Template

### Request

**Method:** `POST`  
**Path:** `/api/v1/templates/{template_id}/render`  
**Content-Type:** `application/json`

**Path Parameters:**
- `template_id` (string, required): The unique identifier of the template

**Headers:**
- `Content-Type: application/json` (set by orchestrator)

**Request Body:**
```json
{
  "language": "en",
  "version": "latest",
  "variables": {
    "name": "John Doe",
    "link": "https://example.com/verify"
  },
  "preview_mode": false
}
```

**Request Body Field Specifications:**

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `language` | string | Yes | Language code for the template |
| `version` | string | No | Template version (default: "latest") |
| `variables` | object | Yes | Key-value pairs for variable substitution |
| `preview_mode` | boolean | Yes | Whether this is a preview (orchestrator sends `false`) |

**Example Request:**
```http
POST /api/v1/templates/welcome_email/render
Content-Type: application/json

{
  "language": "en",
  "version": "latest",
  "variables": {
    "name": "John Doe",
    "link": "https://example.com/verify"
  },
  "preview_mode": false
}
```

### Success Response

**Status Code:** `200 OK`

**Response Body Format:**
The response MUST be a direct JSON object (NOT wrapped in any envelope).

```json
{
  "template_id": "welcome_email",
  "language": "en",
  "version": "1.0.0",
  "rendered": {
    "subject": "Welcome John Doe!",
    "body": {
      "html": "<h1>Welcome John Doe!</h1><p>Click here: https://example.com/verify</p>",
      "text": "Welcome John Doe! Click here: https://example.com/verify"
    }
  },
  "rendered_at": "2025-11-12T12:00:00Z",
  "variables_used": ["name", "link"]
}
```

**Response Field Specifications:**

| Field Name | Type | Required | Description |
|------------|------|----------|-------------|
| `template_id` | string | Yes | The template identifier that was rendered |
| `language` | string | Yes | Language code used for rendering |
| `version` | string | Yes | Version of the template that was rendered |
| `rendered` | object | Yes | The rendered content |
| `rendered.subject` | string | No | Rendered subject (for email templates) |
| `rendered.body` | object | Yes | Rendered body content |
| `rendered.body.html` | string | No | Rendered HTML body |
| `rendered.body.text` | string | Yes | Rendered plain text body |
| `rendered_at` | string (ISO 8601) | Yes | Timestamp when rendering occurred |
| `variables_used` | array of strings | Yes | List of variable names that were actually used in rendering |

**Important Notes:**
- Field names MUST use **snake_case** (e.g., `template_id`, `rendered_at`, `variables_used`)
- The response MUST be a flat JSON object at the root level (no wrapper)
- `rendered.subject` is optional and only present for email templates
- `rendered.body.html` is optional, but `rendered.body.text` is required
- `rendered_at` MUST be in ISO 8601 format
- `variables_used` should only include variables that were actually substituted in the template
- All variable placeholders (`{{variable_name}}`) MUST be replaced with actual values

---

## Error Responses

The orchestrator handles different error scenarios based on HTTP status codes. Error responses can be in any format, but the orchestrator will log the response body for debugging purposes.

### Non-Retryable Errors (4xx - Client Errors)

These errors indicate issues with the request that won't be resolved by retrying.

**Status Codes:**
- `400 Bad Request` - Invalid request format, missing required fields, or invalid variable types
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Template does not exist or language not available
- `422 Unprocessable Entity` - Validation errors (e.g., missing required variables, invalid variable values)

**Behavior:**
- Orchestrator will NOT retry these requests
- Error will be logged and returned to the caller
- Circuit breaker will NOT count these as failures

**Example 404 Response (Template Not Found):**
```json
{
  "error": {
    "code": "TEMPLATE_NOT_FOUND",
    "message": "Template with ID welcome_email does not exist",
    "details": {
      "template_id": "welcome_email",
      "language": "en"
    }
  }
}
```

**Example 422 Response (Missing Required Variables):**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Missing required variables",
    "details": {
      "missing_variables": ["name", "link"],
      "template_id": "welcome_email"
    }
  }
}
```

### Retryable Errors (5xx - Server Errors)

These errors indicate temporary server issues that may be resolved by retrying.

**Status Codes:**
- `500 Internal Server Error` - Server error during template retrieval or rendering
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
```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Unable to process template request",
    "details": {
      "error": "Database connection timeout"
    }
  }
}
```

---

## Variable Substitution

### Variable Format
- Variables in templates use double curly braces: `{{variable_name}}`
- Variable names are case-sensitive
- Variable names should match the keys in the `variables` object of the render request

### Variable Types
The orchestrator sends variables as JSON values. The template service should handle:
- **Strings**: `"John Doe"`
- **Numbers**: `123`, `45.67`
- **Booleans**: `true`, `false`
- **Null**: `null`

### Required Variables
- If a template defines required variables, ALL required variables MUST be provided in the render request
- Missing required variables should return `422 Unprocessable Entity`

### Variable Validation
- The template service should validate variable types match the template's variable definitions
- Invalid types should return `400 Bad Request` or `422 Unprocessable Entity`

---

## Timeout Requirements

- **Request Timeout:** The orchestrator uses a default timeout of 10 seconds
- **Response Time:** 
  - Get Template: Should return within 1-2 seconds
  - Render Template: Should return within 2-3 seconds for optimal performance
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
- Cache frequently accessed templates to improve response times

---

## Testing Checklist

Before integration, please verify:

### Get Template Endpoint:
- [ ] Success response returns `200 OK` with correct JSON structure
- [ ] Field names use snake_case (e.g., `template_id`, `created_at`)
- [ ] Response is NOT wrapped in any envelope (no `data`, `success`, or `result` wrapper)
- [ ] All required fields are present in the response
- [ ] `body.text` is always present (even if empty)
- [ ] `body.html` is optional (can be omitted for non-HTML templates)
- [ ] `subject` is optional (only for email templates)
- [ ] Timestamps are in ISO 8601 format
- [ ] `404 Not Found` is returned for non-existent templates
- [ ] `404 Not Found` is returned for unsupported languages
- [ ] `500` or `503` is returned for server errors (not `404` for server issues)

### Render Template Endpoint:
- [ ] Success response returns `200 OK` with correct JSON structure
- [ ] Field names use snake_case (e.g., `rendered_at`, `variables_used`)
- [ ] Response is NOT wrapped in any envelope
- [ ] All variable placeholders are replaced with actual values
- [ ] `rendered.body.text` is always present
- [ ] `rendered.body.html` is optional
- [ ] `rendered.subject` is optional (only for email templates)
- [ ] `rendered_at` is in ISO 8601 format
- [ ] `variables_used` only includes variables that were actually substituted
- [ ] `422 Unprocessable Entity` is returned for missing required variables
- [ ] `400 Bad Request` is returned for invalid variable types
- [ ] `404 Not Found` is returned for non-existent templates
- [ ] `500` or `503` is returned for server errors

---

## Example Integration Test Cases

### Test Case 1: Get Template - Success
```http
GET /api/v1/templates/welcome_email?language=en
```

**Expected Response:**
```json
{
  "template_id": "welcome_email",
  "name": "Welcome Email",
  "version": "1.0.0",
  "language": "en",
  "type": "email",
  "subject": "Welcome {{name}}!",
  "body": {
    "html": "<h1>Welcome {{name}}!</h1>",
    "text": "Welcome {{name}}!"
  },
  "variables": [
    {
      "name": "name",
      "type": "string",
      "required": true,
      "description": "User's name"
    }
  ],
  "metadata": {
    "created_at": "2025-01-01T00:00:00Z",
    "updated_at": "2025-01-01T00:00:00Z",
    "created_by": "admin",
    "tags": ["welcome"]
  }
}
```

### Test Case 2: Render Template - Success
```http
POST /api/v1/templates/welcome_email/render
Content-Type: application/json

{
  "language": "en",
  "version": "latest",
  "variables": {
    "name": "John Doe"
  },
  "preview_mode": false
}
```

**Expected Response:**
```json
{
  "template_id": "welcome_email",
  "language": "en",
  "version": "1.0.0",
  "rendered": {
    "subject": "Welcome John Doe!",
    "body": {
      "html": "<h1>Welcome John Doe!</h1>",
      "text": "Welcome John Doe!"
    }
  },
  "rendered_at": "2025-11-12T12:00:00Z",
  "variables_used": ["name"]
}
```

### Test Case 3: Template Not Found
```http
GET /api/v1/templates/non_existent?language=en
```

**Expected Response:**
```http
Status: 404 Not Found
```

```json
{
  "error": {
    "code": "TEMPLATE_NOT_FOUND",
    "message": "Template with ID non_existent does not exist",
    "details": {
      "template_id": "non_existent",
      "language": "en"
    }
  }
}
```

### Test Case 4: Missing Required Variables
```http
POST /api/v1/templates/welcome_email/render
Content-Type: application/json

{
  "language": "en",
  "version": "latest",
  "variables": {},
  "preview_mode": false
}
```

**Expected Response:**
```http
Status: 422 Unprocessable Entity
```

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Missing required variables",
    "details": {
      "missing_variables": ["name"],
      "template_id": "welcome_email"
    }
  }
}
```

### Test Case 5: Server Error
```http
GET /api/v1/templates/welcome_email?language=en
```

**Expected Response:**
```http
Status: 503 Service Unavailable
```

```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Unable to process template request",
    "details": {
      "error": "Database connection failed"
    }
  }
}
```
