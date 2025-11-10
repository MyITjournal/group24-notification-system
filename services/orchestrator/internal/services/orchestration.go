package services

import (
	"fmt"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/clients"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrchestrationService struct {
	userClient     clients.UserClient
	templateClient clients.TemplateClient
}

func NewOrchestrationService(userClient clients.UserClient, templateClient clients.TemplateClient) *OrchestrationService {
	return &OrchestrationService{
		userClient:     userClient,
		templateClient: templateClient,
	}
}

func (s *OrchestrationService) ProcessNotification(req *models.NotificationRequest) (*models.NotificationResponse, error) {
	notificationID := uuid.New().String()

	logger.Log.Info("Processing notification",
		zap.String("notification_id", notificationID),
		zap.String("user_id", req.UserID),
		zap.String("template_id", req.TemplateID),
		zap.String("channel", req.Channel),
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

	// Step 2: Check if notifications are enabled globally
	if !userPrefs.NotificationEnabled {
		logger.Log.Warn("Notifications disabled for user",
			zap.String("user_id", req.UserID),
		)
		return &models.NotificationResponse{
			NotificationID: notificationID,
			Status:         "skipped",
			Message:        "User has disabled notifications",
			CreatedAt:      time.Now(),
		}, nil
	}

	// Step 3: Check channel-specific preferences
	if err := s.validateChannelPreferences(req.Channel, userPrefs); err != nil {
		logger.Log.Warn("Channel validation failed",
			zap.String("user_id", req.UserID),
			zap.String("channel", req.Channel),
			zap.Error(err),
		)
		return &models.NotificationResponse{
			NotificationID: notificationID,
			Status:         "skipped",
			Message:        err.Error(),
			CreatedAt:      time.Now(),
		}, nil
	}

	// Step 4: Check opt-out status
	optOutStatus, err := s.userClient.GetOptOutStatus(req.UserID)
	if err != nil {
		logger.Log.Error("Failed to check opt-out status",
			zap.String("user_id", req.UserID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to check opt-out status: %w", err)
	}

	if optOutStatus.OptedOut || optOutStatus.Channels[req.Channel] {
		logger.Log.Warn("User has opted out",
			zap.String("user_id", req.UserID),
			zap.String("channel", req.Channel),
		)
		return &models.NotificationResponse{
			NotificationID: notificationID,
			Status:         "skipped",
			Message:        "User has opted out",
			CreatedAt:      time.Now(),
		}, nil
	}

	// Step 5: Get and render template
	rendered, err := s.templateClient.RenderTemplate(
		req.TemplateID,
		userPrefs.Language,
		req.Variables,
	)
	if err != nil {
		logger.Log.Error("Failed to render template",
			zap.String("template_id", req.TemplateID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// Step 6: Create notification object
	notification := &models.Notification{
		ID:         notificationID,
		UserID:     req.UserID,
		TemplateID: req.TemplateID,
		Channel:    req.Channel,
		To:         s.getRecipient(req.Channel, userPrefs),
		Subject:    rendered.Rendered.HTML, // For email, extract subject
		Body:       rendered.Rendered.Text,
		Variables:  req.Variables,
		Priority:   s.getPriority(req.Priority),
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	// Step 7: TODO - Publish to Kafka (will be implemented later)
	logger.Log.Info("Notification ready for delivery",
		zap.String("notification_id", notificationID),
		zap.String("channel", req.Channel),
		zap.String("to", notification.To),
	)

	// For now, simulate successful queueing
	return &models.NotificationResponse{
		NotificationID: notificationID,
		Status:         "queued",
		Message:        "Notification queued for delivery",
		CreatedAt:      time.Now(),
	}, nil
}

func (s *OrchestrationService) validateChannelPreferences(channel string, prefs *models.UserPreferences) error {
	switch channel {
	case "email":
		if !prefs.Channels.Email.Enabled {
			return fmt.Errorf("email notifications disabled")
		}
		if !prefs.Channels.Email.Verified {
			return fmt.Errorf("email address not verified")
		}
	case "push":
		if !prefs.Channels.Push.Enabled {
			return fmt.Errorf("push notifications disabled")
		}
		if len(prefs.Channels.Push.Devices) == 0 {
			return fmt.Errorf("no active devices registered")
		}
	}
	return nil
}

func (s *OrchestrationService) getRecipient(channel string, prefs *models.UserPreferences) string {
	switch channel {
	case "email":
		return prefs.Email
	case "push":
		if len(prefs.Channels.Push.Devices) > 0 {
			return prefs.Channels.Push.Devices[0].Token
		}
	}
	return ""
}

func (s *OrchestrationService) getPriority(priority string) string {
	if priority == "" {
		return "normal"
	}
	return priority
}
