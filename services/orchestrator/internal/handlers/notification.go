package handlers

import (
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
				"code":       "INVALID_REQUEST",
				"message":    "Invalid request payload",
				"details":    gin.H{"validation_error": err.Error()},
				"request_id": requestID,
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
				"code":       "INTERNAL_ERROR",
				"message":    "Failed to process notification",
				"details":    gin.H{"error": err.Error()},
				"request_id": requestID,
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
