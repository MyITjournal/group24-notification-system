package services

import (
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
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

func TestNewIdempotencyService(t *testing.T) {
	ttl := 5 * time.Minute

	service := NewIdempotencyService(&database.RedisClient{Client: nil}, ttl)

	assert.NotNil(t, service)
	assert.Equal(t, ttl, service.ttl)
}

func TestIdempotencyService_GetRedisKey(t *testing.T) {
	service := &IdempotencyService{
		redisClient: &database.RedisClient{Client: nil},
		ttl:         5 * time.Minute,
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple key",
			input:    "test-key",
			expected: "idempotency:test-key",
		},
		{
			name:     "key with special chars",
			input:    "test-key-123",
			expected: "idempotency:test-key-123",
		},
		{
			name:     "empty key",
			input:    "",
			expected: "idempotency:",
		},
		{
			name:     "uuid key",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: "idempotency:550e8400-e29b-41d4-a716-446655440000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getRedisKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
