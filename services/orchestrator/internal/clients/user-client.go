package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	circuitbreaker "github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/circuit-breaker"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/retry"
	"go.uber.org/zap"
)

type userClient struct {
	baseURL        string
	httpClient     *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
	retryConfig    retry.Config
}

type UserClientConfig struct {
	BaseURL               string
	Timeout               time.Duration
	MaxFailures           uint32
	CircuitBreakerTimeout time.Duration
	HalfOpenMax           uint32
	RetryMaxAttempts      int
	RetryInitialDelay     time.Duration
	RetryMaxDelay         time.Duration
}

func NewUserClient(cfg UserClientConfig) UserClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	retryCfg := retry.DefaultConfig()
	if cfg.RetryMaxAttempts > 0 {
		retryCfg.MaxRetries = cfg.RetryMaxAttempts
	}
	if cfg.RetryInitialDelay > 0 {
		retryCfg.InitialDelay = cfg.RetryInitialDelay
	}
	if cfg.RetryMaxDelay > 0 {
		retryCfg.MaxDelay = cfg.RetryMaxDelay
	}

	return &userClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		circuitBreaker: circuitbreaker.New(circuitbreaker.Config{
			Name:        "user-service",
			MaxFailures: cfg.MaxFailures,
			Timeout:     cfg.CircuitBreakerTimeout,
			HalfOpenMax: cfg.HalfOpenMax,
		}),
		retryConfig: retryCfg,
	}
}

func (c *userClient) GetPreferences(userID string) (*models.UserPreferences, error) {
	var prefs *models.UserPreferences

	ctx := context.Background()

	// Wrap circuit breaker execution with retry logic
	err := retry.Retry(ctx, c.retryConfig, func() error {
		return c.circuitBreaker.Execute(func() error {
			url := fmt.Sprintf("%s/api/v1/users/%s/preferences", c.baseURL, userID)

			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("Content-Type", "application/json")

			logger.Log.Debug("Calling user service",
				zap.String("url", url),
				zap.String("user_id", userID),
			)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				logger.Log.Error("User service request failed",
					zap.String("user_id", userID),
					zap.Error(err),
				)
				return fmt.Errorf("user service request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response body: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				logger.Log.Error("User service returned non-200 status",
					zap.Int("status_code", resp.StatusCode),
					zap.String("user_id", userID),
					zap.String("response_body", string(body)),
				)
				
				// Check if status is retryable
				if retry.IsRetryableHTTPStatus(resp.StatusCode) {
					return fmt.Errorf("user service returned retryable status %d: %s", resp.StatusCode, string(body))
				}
				// Non-retryable error (4xx except 429)
				return fmt.Errorf("user service returned non-retryable status %d: %s", resp.StatusCode, string(body))
			}

			var result models.UserPreferences
			if err := json.Unmarshal(body, &result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}

			prefs = &result
			return nil
		})
	})

	if err != nil {
		// Check if it's a circuit breaker error
		if err == circuitbreaker.ErrCircuitOpen {
			logger.Log.Warn("User service circuit breaker is open",
				zap.String("user_id", userID),
			)
			return nil, fmt.Errorf("user service is temporarily unavailable: %w", err)
		}
		if err == circuitbreaker.ErrTooManyRequests {
			logger.Log.Warn("User service circuit breaker: too many requests in half-open state",
				zap.String("user_id", userID),
			)
			return nil, fmt.Errorf("user service is recovering, please retry: %w", err)
		}
		return nil, err
	}

	return prefs, nil
}
