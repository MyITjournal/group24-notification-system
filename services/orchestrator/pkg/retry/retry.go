package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"go.uber.org/zap"
)

// Config holds retry configuration
type Config struct {
	MaxRetries        int           // Maximum number of retry attempts
	InitialDelay      time.Duration // Initial delay before first retry
	MaxDelay          time.Duration // Maximum delay between retries
	BackoffMultiplier float64       // Multiplier for exponential backoff (default: 2.0)
}

// DefaultConfig returns a default retry configuration
func DefaultConfig() Config {
	return Config{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          5 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// Retry executes a function with retry logic and exponential backoff
func Retry(ctx context.Context, cfg Config, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		if attempt > 0 {
			// Calculate delay with exponential backoff
			delay := calculateDelay(cfg, attempt)

			logger.Log.Debug("Retrying operation",
				zap.Int("attempt", attempt),
				zap.Duration("delay", delay),
				zap.Error(lastErr),
			)

			// Wait before retry, respecting context cancellation
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled: %w", ctx.Err())
			case <-time.After(delay):
				// Continue with retry
			}
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				logger.Log.Info("Operation succeeded after retry",
					zap.Int("attempt", attempt),
				)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			logger.Log.Debug("Error is not retryable, stopping retries",
				zap.Error(err),
			)
			return err
		}

		// If this was the last attempt, return the error
		if attempt == cfg.MaxRetries {
			logger.Log.Warn("Max retries reached, operation failed",
				zap.Int("max_retries", cfg.MaxRetries),
				zap.Error(err),
			)
			return fmt.Errorf("max retries (%d) exceeded: %w", cfg.MaxRetries, err)
		}
	}

	return lastErr
}

// calculateDelay calculates the delay for exponential backoff
func calculateDelay(cfg Config, attempt int) time.Duration {
	// Calculate exponential delay: initialDelay * (multiplier ^ (attempt - 1))
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.BackoffMultiplier, float64(attempt-1))

	// Convert to duration
	duration := time.Duration(delay)

	// Cap at max delay
	if duration > cfg.MaxDelay {
		duration = cfg.MaxDelay
	}

	return duration
}

// isRetryableError determines if an error should be retried
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for network errors (timeout, connection refused, etc.)
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// Check for connection errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	// Check if error message contains retryable patterns
	errMsg := err.Error()
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"no such host",
		"network is unreachable",
		"timeout",
		"temporary failure",
		"service unavailable",
		"bad gateway",
		"gateway timeout",
		"internal server error",
		"status 5", // 5xx errors
	}

	for _, pattern := range retryablePatterns {
		if contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// isRetryableHTTPStatus checks if an HTTP status code is retryable
func IsRetryableHTTPStatus(statusCode int) bool {
	// Retry on 5xx server errors and 429 (Too Many Requests)
	return statusCode >= 500 || statusCode == http.StatusTooManyRequests
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return false // Empty substring doesn't match
	}
	if len(s) < len(substr) {
		return false
	}
	sLower := strings.ToLower(s)
	substrLower := strings.ToLower(substr)
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}
