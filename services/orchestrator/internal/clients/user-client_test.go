package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
	defer logger.Sync()
	code := m.Run()
	os.Exit(code)
}

func TestNewUserClient(t *testing.T) {
	cfg := UserClientConfig{
		BaseURL:               "http://localhost:8080",
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      3,
		RetryInitialDelay:     100 * time.Millisecond,
		RetryMaxDelay:         5 * time.Second,
	}

	client := NewUserClient(cfg)
	assert.NotNil(t, client)
}

func TestNewUserClient_DefaultTimeout(t *testing.T) {
	cfg := UserClientConfig{
		BaseURL: "http://localhost:8080",
		// Timeout is 0, should default to 10 seconds
	}

	client := NewUserClient(cfg)
	assert.NotNil(t, client)
}

func TestUserClient_GetPreferences_Success(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/users/user-123/preferences", r.URL.Path)

		prefs := models.UserPreferences{
			Email: true,
			Push:  false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prefs)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0, // Disable retries for faster tests
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.True(t, prefs.Email)
	assert.False(t, prefs.Push)
}

func TestUserClient_GetPreferences_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "user not found"}`))
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "non-retryable status 404")
}

func TestUserClient_GetPreferences_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0, // Disable retries to test immediate failure
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "retryable status 500")
}

func TestUserClient_GetPreferences_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestUserClient_GetPreferences_NetworkError(t *testing.T) {
	// Use invalid URL to simulate network error
	cfg := UserClientConfig{
		BaseURL:               "http://invalid-host:9999",
		Timeout:               1 * time.Second, // Short timeout for faster test
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0, // Disable retries for faster test
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "user service request failed")
}

func TestUserClient_GetPreferences_Timeout(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               100 * time.Millisecond, // Very short timeout
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
}

func TestUserClient_GetPreferences_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "retryable status 429")
}

func TestUserClient_GetPreferences_AllPreferencesEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefs := models.UserPreferences{
			Email: true,
			Push:  true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prefs)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.True(t, prefs.Email)
	assert.True(t, prefs.Push)
}

func TestUserClient_GetPreferences_AllPreferencesDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefs := models.UserPreferences{
			Email: false,
			Push:  false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prefs)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.False(t, prefs.Email)
	assert.False(t, prefs.Push)
}

func TestUserClient_GetPreferences_WithRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			// First attempt fails with retryable error
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second attempt succeeds
		prefs := models.UserPreferences{
			Email: true,
			Push:  false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(prefs)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      2, // Allow 2 retries
		RetryInitialDelay:     10 * time.Millisecond,
		RetryMaxDelay:         100 * time.Millisecond,
	}

	client := NewUserClient(cfg)
	prefs, err := client.GetPreferences("user-123")

	assert.NoError(t, err)
	assert.NotNil(t, prefs)
	assert.Equal(t, 2, attemptCount) // Should have retried once
}

func TestUserClient_GetPreferences_CircuitBreakerOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := UserClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           2, // Low threshold to trigger circuit breaker quickly
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           1,
		RetryMaxAttempts:      0,
	}

	client := NewUserClient(cfg)

	// Trigger enough failures to open circuit breaker
	for i := 0; i < 3; i++ {
		_, _ = client.GetPreferences("user-123")
	}

	// Now circuit should be open
	prefs, err := client.GetPreferences("user-123")

	assert.Error(t, err)
	assert.Nil(t, prefs)
	assert.Contains(t, err.Error(), "temporarily unavailable")
}
