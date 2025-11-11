package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/services"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	orchestrationService *services.OrchestrationService
}

func NewNotificationHandler(orchestrationService *services.OrchestrationService) *NotificationHandler {
	return &NotificationHandler{
		orchestrationService: orchestrationService,
	}
}

func (h *NotificationHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	startTime := time.Now()

	var req models.NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("Invalid request payload",
			zap.String("request_id", requestID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"success": false,
				"message": "Invalid request payload",
				"error":   gin.H{"validation_error": err.Error()},
			},
		})
		return
	}

	// Process the notification
	response, err := h.orchestrationService.ProcessNotification(&req)
	if err != nil {
		logger.Log.Error("Failed to process notification",
			zap.String("request_id", requestID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"success": false,
				"message": "Failed to process notification",
				"error":   gin.H{"internal server error": err.Error()},
			},
		})
		return
	}

	duration := time.Since(startTime)
	c.Header("X-Response-Time", duration.String())

	statusCode := http.StatusCreated
	if response.Status == "skipped" {
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, response)
}

// UpdateStatus handles POST /api/v1/notifications/:id/status for updating notification status.
func (h *NotificationHandler) UpdateStatus(c *gin.Context) {
	notificationID := c.Param("id")
	requestID, _ := c.Get("request_id")

	// Define a struct to bind the status update payload
	var req struct {
		NotificationID string    `json:"notification_id" binding:"required"`
		Status         string    `json:"status" binding:"required,oneof=delivered pending failed"`
		Timestamp      time.Time `json:"timestamp"`
		Error          string    `json:"error,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("Invalid status update payload",
			zap.String("request_id", requestID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"success": false,
				"message": "Invalid request payload for user creation",
				"error":   gin.H{"validation_error": err.Error()},
			},
		})
		return
	}

	// Basic path validation
	if req.NotificationID != notificationID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"success": false,
				"message": "Notification ID in path does not match payload ID",
				"error":   gin.H{"validation error": req.Error},
			},
		})
		return
	}

	logger.Log.Info("Notification status updated",
		zap.String("notification_id", notificationID),
		zap.String("status", req.Status),
		zap.String("request_id", requestID.(string)),
	)

	// TODO: Implement logic to update status in persistent storage (e.g., database)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Status update received for notification %s: %s", notificationID, req.Status),
		"data": gin.H{
			"notification_id": notificationID,
			"status":          "received",
			"created_at":      time.Now(),
		},
	})
}
