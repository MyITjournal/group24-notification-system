package mocks

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type UserServiceMock struct {
	// Configurable behavior
	simulateDelay    bool
	simulateErrors   bool
	errorRate        float64 // 0.0 to 1.0
	delayMin         time.Duration
	delayMax         time.Duration
	requestCount     int
	failureThreshold int // Fail after N successful requests
}

// MockBehaviorConfig allows configuring mock behavior
type MockBehaviorConfig struct {
	SimulateDelay    bool
	SimulateErrors   bool
	ErrorRate        float64
	DelayMin         time.Duration
	DelayMax         time.Duration
	FailureThreshold int // Fail after N requests (for circuit breaker testing)
}

func NewUserServiceMock() *UserServiceMock {
	// Read configuration from environment variables
	simulateDelay := os.Getenv("MOCK_SIMULATE_DELAY") == "true"
	simulateErrors := os.Getenv("MOCK_SIMULATE_ERRORS") == "true"

	var errorRate float64 = 0.1 // 10% error rate by default
	if rate := os.Getenv("MOCK_ERROR_RATE"); rate != "" {
		if _, err := fmt.Sscanf(rate, "%f", &errorRate); err != nil {
			errorRate = 0.1
		}
	}

	delayMin := 50 * time.Millisecond
	delayMax := 200 * time.Millisecond
	if min := os.Getenv("MOCK_DELAY_MIN"); min != "" {
		if d, err := time.ParseDuration(min); err == nil {
			delayMin = d
		}
	}
	if max := os.Getenv("MOCK_DELAY_MAX"); max != "" {
		if d, err := time.ParseDuration(max); err == nil {
			delayMax = d
		}
	}

	return &UserServiceMock{
		simulateDelay:    simulateDelay,
		simulateErrors:   simulateErrors,
		errorRate:        errorRate,
		delayMin:         delayMin,
		delayMax:         delayMax,
		requestCount:     0,
		failureThreshold: 0, // 0 = disabled
	}
}

// NewUserServiceMockWithConfig creates a mock with custom configuration
func NewUserServiceMockWithConfig(cfg MockBehaviorConfig) *UserServiceMock {
	return &UserServiceMock{
		simulateDelay:    cfg.SimulateDelay,
		simulateErrors:   cfg.SimulateErrors,
		errorRate:        cfg.ErrorRate,
		delayMin:         cfg.DelayMin,
		delayMax:         cfg.DelayMax,
		requestCount:     0,
		failureThreshold: cfg.FailureThreshold,
	}
}

func (m *UserServiceMock) GetPreferences(userID string) (*models.UserPreferences, error) {
	m.requestCount++

	// Simulate network delay (realistic latency)
	if m.simulateDelay {
		delay := m.delayMin + time.Duration(rand.Int63n(int64(m.delayMax-m.delayMin)))
		time.Sleep(delay)
	}

	// Simulate failure threshold (for circuit breaker testing)
	if m.failureThreshold > 0 && m.requestCount > m.failureThreshold {
		return nil, fmt.Errorf("user service unavailable: service overloaded")
	}

	// Simulate random errors based on error rate
	if m.simulateErrors && rand.Float64() < m.errorRate {
		return m.simulateError(userID)
	}

	// Simulate specific error scenarios based on user ID patterns
	if strings.HasPrefix(userID, "error_") {
		return m.simulateError(userID)
	}

	// Simulate user not found
	if strings.HasPrefix(userID, "notfound_") || userID == "usr_notfound" {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	// Simulate timeout scenario
	if strings.HasPrefix(userID, "timeout_") {
		time.Sleep(5 * time.Second) // Longer than typical timeout
		return nil, fmt.Errorf("request timeout: user service did not respond")
	}

	// Simulate service unavailable
	if strings.HasPrefix(userID, "unavailable_") {
		return nil, fmt.Errorf("user service unavailable: service temporarily down")
	}

	// Simulate invalid response (malformed data)
	if strings.HasPrefix(userID, "invalid_") {
		// This would cause JSON unmarshaling to fail in real client
		return nil, fmt.Errorf("invalid response format from user service")
	}

	// Default: Return realistic user preferences based on user ID
	prefs := m.getRealisticPreferences(userID)
	return prefs, nil
}

func (m *UserServiceMock) simulateError(userID string) (*models.UserPreferences, error) {
	errorTypes := []string{
		"internal_server_error",
		"service_unavailable",
		"timeout",
		"database_error",
		"network_error",
	}

	errorType := errorTypes[rand.Intn(len(errorTypes))]

	switch errorType {
	case "internal_server_error":
		return nil, fmt.Errorf("user service internal server error: %s", userID)
	case "service_unavailable":
		return nil, fmt.Errorf("user service unavailable: %s", userID)
	case "timeout":
		return nil, fmt.Errorf("user service request timeout: %s", userID)
	case "database_error":
		return nil, fmt.Errorf("user service database error: %s", userID)
	case "network_error":
		return nil, fmt.Errorf("user service network error: connection refused")
	default:
		return nil, fmt.Errorf("user service error: %s", userID)
	}
}

func (m *UserServiceMock) getRealisticPreferences(userID string) *models.UserPreferences {
	// Simulate different user preference scenarios
	prefs := &models.UserPreferences{}

	// Users with email disabled
	if strings.Contains(userID, "no_email") || strings.HasPrefix(userID, "usr_noemail_") {
		prefs.Email = false
		prefs.Push = true
		return prefs
	}

	// Users with push disabled
	if strings.Contains(userID, "no_push") || strings.HasPrefix(userID, "usr_nopush_") {
		prefs.Email = true
		prefs.Push = false
		return prefs
	}

	// Users with both disabled
	if strings.Contains(userID, "no_notifications") || strings.HasPrefix(userID, "usr_nonotif_") {
		prefs.Email = false
		prefs.Push = false
		return prefs
	}

	// Users with email only
	if strings.Contains(userID, "email_only") || strings.HasPrefix(userID, "usr_emailonly_") {
		prefs.Email = true
		prefs.Push = false
		return prefs
	}

	// Users with push only
	if strings.Contains(userID, "push_only") || strings.HasPrefix(userID, "usr_pushonly_") {
		prefs.Email = false
		prefs.Push = true
		return prefs
	}

	// Default: Both enabled (most common case)
	prefs.Email = true
	prefs.Push = true
	return prefs
}

// Reset resets the mock state (useful for testing)
func (m *UserServiceMock) Reset() {
	m.requestCount = 0
}

// GetRequestCount returns the number of requests made (for testing)
func (m *UserServiceMock) GetRequestCount() int {
	return m.requestCount
}
