package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// IdempotencyService handles idempotency checks using Redis
type IdempotencyService struct {
	redisClient *database.RedisClient
	ttl         time.Duration
}

// NewIdempotencyService creates a new idempotency service
func NewIdempotencyService(redisClient *database.RedisClient, ttl time.Duration) *IdempotencyService {
	return &IdempotencyService{
		redisClient: redisClient,
		ttl:         ttl,
	}
}

// GetCachedResponse checks if an idempotency key exists in Redis and returns the cached response
func (s *IdempotencyService) GetCachedResponse(ctx context.Context, idempotencyKey string) (*models.NotificationResponse, error) {
	key := s.getRedisKey(idempotencyKey)

	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			// Key doesn't exist, not an error
			return nil, nil
		}
		logger.Log.Error("Failed to get idempotency key from Redis",
			zap.String("idempotency_key", idempotencyKey),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check idempotency key: %w", err)
	}

	// Deserialize the cached response
	var response models.NotificationResponse
	if err := json.Unmarshal([]byte(val), &response); err != nil {
		logger.Log.Error("Failed to unmarshal cached response",
			zap.String("idempotency_key", idempotencyKey),
			zap.Error(err),
		)
		// If unmarshaling fails, treat as if key doesn't exist
		return nil, nil
	}

	logger.Log.Info("Idempotency key found, returning cached response",
		zap.String("idempotency_key", idempotencyKey),
		zap.String("notification_id", response.NotificationID),
	)

	return &response, nil
}

// StoreResponse stores the response in Redis with TTL for idempotency
func (s *IdempotencyService) StoreResponse(ctx context.Context, idempotencyKey string, response *models.NotificationResponse) error {
	key := s.getRedisKey(idempotencyKey)

	// Serialize the response
	data, err := json.Marshal(response)
	if err != nil {
		logger.Log.Error("Failed to marshal response for idempotency",
			zap.String("idempotency_key", idempotencyKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Store in Redis with TTL
	if err := s.redisClient.Set(ctx, key, data, s.ttl).Err(); err != nil {
		logger.Log.Error("Failed to store idempotency key in Redis",
			zap.String("idempotency_key", idempotencyKey),
			zap.Error(err),
		)
		return fmt.Errorf("failed to store idempotency key: %w", err)
	}

	logger.Log.Debug("Stored idempotency key in Redis",
		zap.String("idempotency_key", idempotencyKey),
		zap.String("notification_id", response.NotificationID),
		zap.Duration("ttl", s.ttl),
	)

	return nil
}

// getRedisKey generates the Redis key for an idempotency key
func (s *IdempotencyService) getRedisKey(idempotencyKey string) string {
	return fmt.Sprintf("idempotency:%s", idempotencyKey)
}

