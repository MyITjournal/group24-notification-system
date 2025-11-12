package services

import (
	"context"
	"fmt"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/clients"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/repository"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/kafka"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrchestrationService struct {
	userClient       clients.UserClient
	templateClient   clients.TemplateClient
	kafkaManager     *kafka.Manager
	notificationRepo repository.NotificationRepository
}

func NewOrchestrationService(
	userClient clients.UserClient,
	templateClient clients.TemplateClient,
	kafkaManager *kafka.Manager,
	notificationRepo repository.NotificationRepository,
) *OrchestrationService {
	return &OrchestrationService{
		userClient:       userClient,
		templateClient:   templateClient,
		kafkaManager:     kafkaManager,
		notificationRepo: notificationRepo,
	}
}

func (s *OrchestrationService) ProcessNotification(req *models.NotificationRequest) (*models.NotificationResponse, error) {
	notificationID := uuid.New().String()
	ctx := context.Background()

	logger.Log.Info("Processing notification",
		zap.String("notification_id", notificationID),
		zap.String("user_id", req.UserID),
		zap.String("template_code", req.TemplateCode),
		zap.String("notification_type", string(req.NotificationType)),
	)

	// Step 1: Get user preferences
	userPrefs, err := s.userClient.GetPreferences(req.UserID)
	if err != nil {
		logger.Log.Error("Failed to get user preferences",
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	// Step 2: Validate channel preferences
	if err := s.validateChannelPreferences(req.NotificationType, userPrefs); err != nil {
		logger.Log.Warn("Channel validation failed",
			zap.String("user_id", req.UserID),
			zap.String("notification_type", string(req.NotificationType)),
			zap.Error(err),
		)

		// Persist failed notification for audit trail
		errorMsg := err.Error()
		notificationRecord := &models.NotificationRecord{
			ID:               notificationID,
			UserID:           req.UserID,
			TemplateCode:     req.TemplateCode,
			NotificationType: string(req.NotificationType),
			Status:           models.StatusFailed,
			Priority:         s.getPriority(req.Priority),
			Variables:        models.JSONB(req.Variables),
			ScheduledFor:     req.ScheduledFor,
			ErrorMessage:     &errorMsg,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if req.Metadata != nil {
			metadata := models.JSONB(req.Metadata)
			notificationRecord.Metadata = &metadata
		}

		if persistErr := s.notificationRepo.Create(ctx, notificationRecord); persistErr != nil {
			logger.Log.Error("Failed to persist failed notification record",
				zap.String("notification_id", notificationID),
				zap.Error(persistErr),
			)
		}

		return &models.NotificationResponse{
			NotificationID: notificationID,
			Status:         models.StatusFailed,
			Timestamp:      time.Now(),
			Error:          err.Error(),
		}, nil
	}

	// Step 3: Render template
	rendered, err := s.templateClient.RenderTemplate(
		req.TemplateCode,
		"en", // Default language
		req.Variables,
	)
	if err != nil {
		logger.Log.Error("Failed to render template",
			zap.String("template_code", req.TemplateCode),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Step 4: Create notification record
	notificationRecord := &models.NotificationRecord{
		ID:               notificationID,
		UserID:           req.UserID,
		TemplateCode:     req.TemplateCode,
		NotificationType: string(req.NotificationType),
		Status:           models.StatusPending,
		Priority:         s.getPriority(req.Priority),
		Variables:        models.JSONB(req.Variables),
		ScheduledFor:     req.ScheduledFor,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if req.Metadata != nil {
		metadata := models.JSONB(req.Metadata)
		notificationRecord.Metadata = &metadata
	}

	// Persist notification record (even if Kafka publish fails, we want audit trail)
	if err := s.notificationRepo.Create(ctx, notificationRecord); err != nil {
		logger.Log.Error("Failed to persist notification record",
			zap.String("notification_id", notificationID),
			zap.Error(err),
		)
		// Continue processing even if persistence fails, but log the error
	}

	// Step 5: Create and publish Kafka payload
	payload := s.createKafkaPayload(notificationID, req, rendered)
	if err := s.publishToKafka(ctx, req.NotificationType, notificationID, payload); err != nil {
		logger.Log.Error("Failed to publish to Kafka",
			zap.String("notification_id", notificationID),
			zap.Error(err),
		)
		// Update status to failed if Kafka publish fails
		if updateErr := s.notificationRepo.UpdateStatus(ctx, notificationID, models.StatusFailed, err.Error()); updateErr != nil {
			logger.Log.Error("Failed to update notification status after Kafka error",
				zap.String("notification_id", notificationID),
				zap.Error(updateErr),
			)
		}
		return nil, fmt.Errorf("failed to queue notification: %w", err)
	}

	logger.Log.Info("Notification queued successfully",
		zap.String("notification_id", notificationID),
		zap.String("notification_type", string(req.NotificationType)),
	)

	return &models.NotificationResponse{
		NotificationID: notificationID,
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}, nil
}

// createKafkaPayload constructs the payload for Kafka based on notification type
func (s *OrchestrationService) createKafkaPayload(
	notificationID string,
	req *models.NotificationRequest,
	rendered *models.RenderResponse,
) *models.KafkaNotificationPayload {
	payload := &models.KafkaNotificationPayload{
		NotificationID:   notificationID,
		NotificationType: string(req.NotificationType),
		UserID:           req.UserID,
		TemplateCode:     req.TemplateCode,
		Priority:         s.getPriority(req.Priority),
		Metadata:         req.Metadata,
		CreatedAt:        time.Now(),
	}

	// Set rendered content based on notification type
	switch req.NotificationType {
	case models.NotificationEmail:
		payload.Subject = rendered.Rendered.Subject
		payload.Body = rendered.Rendered.Body.HTML
		payload.TextBody = rendered.Rendered.Body.Text
	case models.NotificationPush:
		// For push notifications, use the text body as the message
		payload.Body = rendered.Rendered.Body.Text
		// Subject can be used as the push notification title
		if rendered.Rendered.Subject != "" {
			payload.Subject = rendered.Rendered.Subject
		}
	}

	return payload
}

// publishToKafka sends the notification to the appropriate Kafka topic
func (s *OrchestrationService) publishToKafka(
	ctx context.Context,
	notificationType models.NotificationType,
	key string,
	payload *models.KafkaNotificationPayload,
) error {
	return s.kafkaManager.PublishByType(
		ctx,
		string(notificationType),
		key,
		payload,
	)
}

func (s *OrchestrationService) validateChannelPreferences(
	notificationType models.NotificationType,
	prefs *models.UserPreferences,
) error {
	switch notificationType {
	case models.NotificationEmail:
		if !prefs.Email {
			return fmt.Errorf("email notifications disabled")
		}
	case models.NotificationPush:
		if !prefs.Push {
			return fmt.Errorf("push notifications disabled")
		}
	default:
		return fmt.Errorf("unknown notification type: %s", notificationType)
	}
	return nil
}

func (s *OrchestrationService) getPriority(priority int) string {
	// Map integer priority to string values
	// 0 or unset = normal, 1 = low, 2 = normal, 3 = high, 4 = urgent
	if priority == 0 {
		return "normal"
	}

	priorityMap := map[int]string{
		1: "low",
		2: "normal",
		3: "high",
		4: "urgent",
	}

	if val, ok := priorityMap[priority]; ok {
		return val
	}

	// Default to normal for unknown values
	return "normal"
}

// UpdateNotificationStatus updates the status of a notification in the database
func (s *OrchestrationService) UpdateNotificationStatus(ctx context.Context, notificationID string, status models.NotificationStatus, errorMsg string) error {
	return s.notificationRepo.UpdateStatus(ctx, notificationID, status, errorMsg)
}
