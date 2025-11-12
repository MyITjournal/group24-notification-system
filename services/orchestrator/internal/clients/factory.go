package clients

import (
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/config"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/mocks"
)

// NewUserClientFromConfig creates a UserClient based on configuration
// If UseMockServices is true, returns a mock; otherwise returns a real client
func NewUserClientFromConfig(cfg config.ServicesConfig) UserClient {
	if cfg.UseMockServices {
		return mocks.NewUserServiceMock()
	}

	return NewUserClient(UserClientConfig{
		BaseURL:               cfg.UserService.BaseURL,
		Timeout:               cfg.UserService.Timeout,
		RetryMaxAttempts:      cfg.UserService.RetryMaxAttempts,
		RetryInitialDelay:     cfg.UserService.RetryInitialDelay,
		RetryMaxDelay:         cfg.UserService.RetryMaxDelay,
		MaxFailures:           5,  // Default circuit breaker settings
		CircuitBreakerTimeout: 60, // seconds
		HalfOpenMax:           3,
	})
}

// NewTemplateClientFromConfig creates a TemplateClient based on configuration
// If UseMockServices is true, returns a mock; otherwise returns a real client
func NewTemplateClientFromConfig(cfg config.ServicesConfig) TemplateClient {
	if cfg.UseMockServices {
		return mocks.NewTemplateServiceMock()
	}

	return NewTemplateClient(TemplateClientConfig{
		BaseURL:               cfg.TemplateService.BaseURL,
		Timeout:               cfg.TemplateService.Timeout,
		RetryMaxAttempts:      cfg.TemplateService.RetryMaxAttempts,
		RetryInitialDelay:     cfg.TemplateService.RetryInitialDelay,
		RetryMaxDelay:         cfg.TemplateService.RetryMaxDelay,
		MaxFailures:           5,  // Default circuit breaker settings
		CircuitBreakerTimeout: 60, // seconds
		HalfOpenMax:           3,
	})
}
