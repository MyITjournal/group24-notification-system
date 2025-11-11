package kafka

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// Manager handles multiple Kafka producers for different topics
type Manager struct {
	emailProducer *Producer
	pushProducer  *Producer
	logger        *zap.Logger
}

type ManagerConfig struct {
	Brokers    []string
	EmailTopic string
	PushTopic  string
	Logger     *zap.Logger
}

func NewManager(cfg ManagerConfig) (*Manager, error) {
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker is required")
	}

	emailProducer := NewProducer(ProducerConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.EmailTopic,
		Logger:  cfg.Logger,
	})

	pushProducer := NewProducer(ProducerConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.PushTopic,
		Logger:  cfg.Logger,
	})

	return &Manager{
		emailProducer: emailProducer,
		pushProducer:  pushProducer,
		logger:        cfg.Logger,
	}, nil
}

// PublishEmail publishes a message to the email queue
func (m *Manager) PublishEmail(ctx context.Context, notificationID string, payload interface{}) error {
	m.logger.Info("Publishing to email queue",
		zap.String("notification_id", notificationID),
	)
	return m.emailProducer.Publish(ctx, notificationID, payload)
}

// PublishPush publishes a message to the push notification queue
func (m *Manager) PublishPush(ctx context.Context, notificationID string, payload interface{}) error {
	m.logger.Info("Publishing to push queue",
		zap.String("notification_id", notificationID),
	)
	return m.pushProducer.Publish(ctx, notificationID, payload)
}

// PublishByType routes to the correct queue based on notification type
func (m *Manager) PublishByType(ctx context.Context, notificationType, notificationID string, payload interface{}) error {
	switch notificationType {
	case "email":
		return m.PublishEmail(ctx, notificationID, payload)
	case "push":
		return m.PublishPush(ctx, notificationID, payload)
	default:
		return fmt.Errorf("unsupported notification type: %s", notificationType)
	}
}

// Close closes all producers
func (m *Manager) Close() error {
	m.logger.Info("Closing Kafka manager")

	if err := m.emailProducer.Close(); err != nil {
		m.logger.Error("Failed to close email producer", zap.Error(err))
		return err
	}

	if err := m.pushProducer.Close(); err != nil {
		m.logger.Error("Failed to close push producer", zap.Error(err))
		return err
	}

	return nil
}

// HealthCheck verifies connectivity to Kafka brokers
func (m *Manager) HealthCheck() error {
	// Check email producer stats
	emailStats := m.emailProducer.Stats()
	if emailStats.Errors > 0 {
		m.logger.Warn("Email producer has errors",
			zap.Int64("errors", emailStats.Errors),
		)
	}

	// Check push producer stats
	pushStats := m.pushProducer.Stats()
	if pushStats.Errors > 0 {
		m.logger.Warn("Push producer has errors",
			zap.Int64("errors", pushStats.Errors),
		)
	}

	return nil
}
