package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

type ProducerConfig struct {
	Brokers []string
	Topic   string
	Logger  *zap.Logger
}

type Message struct {
	Key   string
	Value interface{}
}

func NewProducer(cfg ProducerConfig) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  3,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
		RequiredAcks: kafka.RequireOne,
		Async:        false, // Synchronous for reliability
	}

	return &Producer{
		writer: writer,
		logger: cfg.Logger,
	}
}

// Publish sends a message to Kafka with retries
func (p *Producer) Publish(ctx context.Context, key string, value interface{}) error {
	valueBytes, err := json.Marshal(value)
	if err != nil {
		p.logger.Error("Failed to marshal message",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: valueBytes,
		Time:  time.Now(),
	}

	p.logger.Debug("Publishing message to Kafka",
		zap.String("topic", p.writer.Topic),
		zap.String("key", key),
	)

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.Error("Failed to publish message",
			zap.String("topic", p.writer.Topic),
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.Info("Message published successfully",
		zap.String("topic", p.writer.Topic),
		zap.String("key", key),
	)

	return nil
}

// PublishBatch sends multiple messages in a batch
func (p *Producer) PublishBatch(ctx context.Context, messages []Message) error {
	kafkaMessages := make([]kafka.Message, len(messages))

	for i, msg := range messages {
		valueBytes, err := json.Marshal(msg.Value)
		if err != nil {
			p.logger.Error("Failed to marshal batch message",
				zap.Int("index", i),
				zap.Error(err),
			)
			return fmt.Errorf("failed to marshal batch message at index %d: %w", i, err)
		}

		kafkaMessages[i] = kafka.Message{
			Key:   []byte(msg.Key),
			Value: valueBytes,
			Time:  time.Now(),
		}
	}

	err := p.writer.WriteMessages(ctx, kafkaMessages...)
	if err != nil {
		p.logger.Error("Failed to publish batch",
			zap.Int("count", len(messages)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to publish batch: %w", err)
	}

	p.logger.Info("Batch published successfully",
		zap.Int("count", len(messages)),
	)

	return nil
}

// Close gracefully shuts down the producer
func (p *Producer) Close() error {
	p.logger.Info("Closing Kafka producer")
	return p.writer.Close()
}

// Stats returns producer statistics
func (p *Producer) Stats() kafka.WriterStats {
	return p.writer.Stats()
}
