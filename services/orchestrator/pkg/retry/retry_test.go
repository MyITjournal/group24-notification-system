package retry

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

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

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 100*time.Millisecond, cfg.InitialDelay)
	assert.Equal(t, 5*time.Second, cfg.MaxDelay)
	assert.Equal(t, 2.0, cfg.BackoffMultiplier)
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	cfg := DefaultConfig()
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 3
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		if callCount < 3 {
			return errors.New("timeout error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}

func TestRetry_MaxRetriesExceeded(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 2
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	expectedErr := errors.New("service unavailable")
	err := Retry(ctx, cfg, func() error {
		callCount++
		return expectedErr
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (2) exceeded")
	assert.Equal(t, 3, callCount) // Initial attempt + 2 retries
}

func TestRetry_NonRetryableError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	ctx := context.Background()

	callCount := 0
	expectedErr := errors.New("bad request: invalid input")
	err := Retry(ctx, cfg, func() error {
		callCount++
		return expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 1, callCount) // Should not retry non-retryable errors
}

func TestRetry_ContextCancellation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	cfg.InitialDelay = 50 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		if callCount == 1 {
			// Cancel context before retry
			cancel()
		}
		return errors.New("timeout error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	assert.Equal(t, 1, callCount)
}

func TestRetry_ContextTimeout(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	cfg.InitialDelay = 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		return errors.New("timeout error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 3
	cfg.InitialDelay = 10 * time.Millisecond
	cfg.BackoffMultiplier = 2.0
	ctx := context.Background()

	delays := []time.Duration{}
	lastCallTime := time.Now()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		now := time.Now()
		if callCount > 0 {
			delays = append(delays, now.Sub(lastCallTime))
		}
		lastCallTime = now
		callCount++
		if callCount < 4 {
			return errors.New("timeout error")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 4, callCount)
	assert.Len(t, delays, 3)

	// Verify exponential backoff (with some tolerance for timing)
	assert.GreaterOrEqual(t, delays[0], 8*time.Millisecond) // ~10ms
	assert.LessOrEqual(t, delays[0], 20*time.Millisecond)
	assert.GreaterOrEqual(t, delays[1], 18*time.Millisecond) // ~20ms
	assert.LessOrEqual(t, delays[1], 30*time.Millisecond)
	assert.GreaterOrEqual(t, delays[2], 38*time.Millisecond) // ~40ms
	assert.LessOrEqual(t, delays[2], 50*time.Millisecond)
}

func TestRetry_MaxDelayCapping(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 3
	cfg.InitialDelay = 100 * time.Millisecond
	cfg.MaxDelay = 150 * time.Millisecond
	cfg.BackoffMultiplier = 2.0
	ctx := context.Background()

	delays := []time.Duration{}
	lastCallTime := time.Now()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		now := time.Now()
		if callCount > 0 {
			delays = append(delays, now.Sub(lastCallTime))
		}
		lastCallTime = now
		callCount++
		if callCount < 4 {
			return errors.New("timeout error")
		}
		return nil
	})

	assert.NoError(t, err)
	// Verify delays are capped at MaxDelay
	for _, delay := range delays {
		assert.LessOrEqual(t, delay, 200*time.Millisecond) // Allow some tolerance
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   bool
	}{
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"429 Too Many Requests", 429, true},
		{"200 OK", 200, false},
		{"400 Bad Request", 400, false},
		{"404 Not Found", 404, false},
		{"401 Unauthorized", 401, false},
		{"499 Client Closed Request", 499, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableHTTPStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryableError_NetworkErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"Timeout error", &net.DNSError{Err: "timeout", IsTimeout: true}, true},
		{"Temporary error", &net.DNSError{Err: "temporary", IsTemporary: true}, true},
		{"DNS error", &net.DNSError{Err: "no such host"}, true},
		{"OpError", &net.OpError{Op: "read", Err: errors.New("connection refused")}, true},
		{"Connection refused in message", errors.New("connection refused"), true},
		{"Connection reset in message", errors.New("connection reset"), true},
		{"Connection timeout in message", errors.New("connection timeout"), true},
		{"No such host in message", errors.New("no such host"), true},
		{"Network unreachable in message", errors.New("network is unreachable"), true},
		{"Timeout in message", errors.New("timeout"), true},
		{"Temporary failure in message", errors.New("temporary failure"), true},
		{"Service unavailable in message", errors.New("service unavailable"), true},
		{"Bad gateway in message", errors.New("bad gateway"), true},
		{"Gateway timeout in message", errors.New("gateway timeout"), true},
		{"Internal server error in message", errors.New("internal server error"), true},
		{"Status 500 in message", errors.New("status 500"), true},
		{"Status 503 in message", errors.New("status 503"), true},
		{"Non-retryable error", errors.New("bad request: invalid input"), false},
		{"Nil error", nil, false},
		{"Validation error", errors.New("validation failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result, "Error: %v", tt.err)
		})
	}
}

func TestIsRetryableError_CaseInsensitive(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{errors.New("CONNECTION REFUSED"), true},
		{errors.New("Timeout"), true},
		{errors.New("SERVICE UNAVAILABLE"), true},
		{errors.New("Bad Request"), false},
	}

	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			result := isRetryableError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 100 * time.Millisecond
	cfg.BackoffMultiplier = 2.0
	cfg.MaxDelay = 1 * time.Second

	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{"First retry (attempt 1)", 1, 100 * time.Millisecond},  // 100ms * 2^0
		{"Second retry (attempt 2)", 2, 200 * time.Millisecond}, // 100ms * 2^1
		{"Third retry (attempt 3)", 3, 400 * time.Millisecond},  // 100ms * 2^2
		{"Fourth retry (attempt 4)", 4, 800 * time.Millisecond}, // 100ms * 2^3
		{"Fifth retry (attempt 5)", 5, 1 * time.Second},         // Capped at MaxDelay
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := calculateDelay(cfg, tt.attempt)
			assert.Equal(t, tt.expected, delay)
		})
	}
}

func TestCalculateDelay_CustomMultiplier(t *testing.T) {
	cfg := DefaultConfig()
	cfg.InitialDelay = 50 * time.Millisecond
	cfg.BackoffMultiplier = 1.5
	cfg.MaxDelay = 500 * time.Millisecond

	delay1 := calculateDelay(cfg, 1)
	assert.Equal(t, 50*time.Millisecond, delay1)

	delay2 := calculateDelay(cfg, 2)
	assert.Equal(t, 75*time.Millisecond, delay2) // 50 * 1.5

	delay3 := calculateDelay(cfg, 3)
	// 50 * 1.5^2 = 112.5ms, but duration is rounded, so check with tolerance
	assert.GreaterOrEqual(t, delay3, 112*time.Millisecond)
	assert.LessOrEqual(t, delay3, 113*time.Millisecond)
}

func TestRetry_ZeroMaxRetries(t *testing.T) {
	cfg := Config{
		MaxRetries:        0,
		InitialDelay:      10 * time.Millisecond,
		MaxDelay:          100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Only initial attempt, no retries
}

func TestRetry_MixedRetryableAndNonRetryable(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 3
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("timeout") // Retryable
		}
		return errors.New("bad request: invalid") // Non-retryable
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bad request")
	assert.Equal(t, 2, callCount) // Initial + 1 retry, then stops
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"connection refused", "connection", true},
		{"connection refused", "refused", true},
		{"connection refused", "CONNECTION", true}, // Case insensitive
		{"timeout error", "timeout", true},
		{"short", "very long substring", false},
		{"", "substring", false},
		{"substring", "", false}, // Empty substring doesn't match
		{"exact match", "exact match", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+" contains "+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry_WithHTTPStatusError(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 2
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		return errors.New("status 503 service unavailable")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries")
	assert.Equal(t, 3, callCount) // Initial + 2 retries
}

func TestRetry_ImmediateSuccessAfterFailure(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		if callCount == 1 {
			return errors.New("timeout")
		}
		return nil // Success on retry
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, callCount)
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 2
	cfg.InitialDelay = 10 * time.Millisecond
	ctx := context.Background()

	callCount := 0
	expectedErr := errors.New("persistent timeout")
	err := Retry(ctx, cfg, func() error {
		callCount++
		return expectedErr
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "max retries (2) exceeded")
	assert.Equal(t, 3, callCount) // Initial + 2 retries
}

func TestRetry_ContextAlreadyCancelled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRetries = 5
	cfg.InitialDelay = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	callCount := 0
	err := Retry(ctx, cfg, func() error {
		callCount++
		return errors.New("error")
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cancelled")
	// Context is checked before calling operation, so operation should not be called
	assert.Equal(t, 0, callCount)
}
