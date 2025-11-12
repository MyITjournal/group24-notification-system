package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OrchestrationServiceInterface defines the interface for orchestration operations
type OrchestrationServiceInterface interface {
	ProcessNotification(req *models.NotificationRequest) (*models.NotificationResponse, error)
	UpdateNotificationStatus(ctx context.Context, notificationID string, status models.NotificationStatus, errorMsg string) error
}

// IdempotencyServiceInterface defines the interface for idempotency operations
type IdempotencyServiceInterface interface {
	GetCachedResponse(ctx context.Context, key string) (*models.NotificationResponse, error)
	StoreResponse(ctx context.Context, key string, response *models.NotificationResponse) error
}

type NotificationHandler struct {
	orchestrationService OrchestrationServiceInterface
	idempotencyService   IdempotencyServiceInterface
}

func NewNotificationHandler(
	orchestrationService OrchestrationServiceInterface,
	idempotencyService IdempotencyServiceInterface,
) *NotificationHandler {
	return &NotificationHandler{
		orchestrationService: orchestrationService,
		idempotencyService:   idempotencyService,
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
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	// Check for idempotency - use the RequestID field as the idempotency key
	ctx := context.Background()
	cachedResponse, err := h.idempotencyService.GetCachedResponse(ctx, req.RequestID)
	if err != nil {
		logger.Log.Warn("Failed to check idempotency, proceeding with request",
			zap.String("request_id", requestID.(string)),
			zap.String("idempotency_key", req.RequestID),
			zap.Error(err),
		)
		// Continue processing even if idempotency check fails
	} else if cachedResponse != nil {
		// Return cached response
		logger.Log.Info("Returning cached response for idempotency key",
			zap.String("request_id", requestID.(string)),
			zap.String("idempotency_key", req.RequestID),
			zap.String("notification_id", cachedResponse.NotificationID),
		)
		duration := time.Since(startTime)
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Idempotent-Replay", "true")
		c.JSON(http.StatusOK, models.Response{
			Success: true,
			Message: "Notification retrieved from cache",
			Data:    cachedResponse,
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
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to process notification",
			Error:   err.Error(),
		})
		return
	}

	// Store response in Redis for idempotency (even if it fails, we don't want to block the response)
	if storeErr := h.idempotencyService.StoreResponse(ctx, req.RequestID, response); storeErr != nil {
		logger.Log.Warn("Failed to store idempotency key, response still returned",
			zap.String("request_id", requestID.(string)),
			zap.String("idempotency_key", req.RequestID),
			zap.Error(storeErr),
		)
	}

	duration := time.Since(startTime)
	c.Header("X-Response-Time", duration.String())

	c.JSON(http.StatusCreated, models.Response{
		Success: true,
		Message: "Notification queued successfully",
		Data:    response,
	})
}

// UpdateStatus handles POST /api/v1/{notification_type}/status for updating notification status.
func (h *NotificationHandler) UpdateStatus(c *gin.Context) {
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
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Invalid request payload",
			Error:   err.Error(),
		})
		return
	}

	notificationID := req.NotificationID

	logger.Log.Info("Updating notification status",
		zap.String("notification_id", notificationID),
		zap.String("status", req.Status),
		zap.String("request_id", requestID.(string)),
	)

	// Convert string status to NotificationStatus type
	notificationStatus := models.NotificationStatus(req.Status)
	if notificationStatus != models.StatusPending && notificationStatus != models.StatusDelivered && notificationStatus != models.StatusFailed {
		c.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: "Invalid status value",
			Error:   "status must be one of: pending, delivered, failed",
		})
		return
	}

	// Update status in database
	ctx := context.Background()
	if err := h.orchestrationService.UpdateNotificationStatus(ctx, notificationID, notificationStatus, req.Error); err != nil {
		logger.Log.Error("Failed to update notification status",
			zap.String("notification_id", notificationID),
			zap.String("request_id", requestID.(string)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: "Failed to update notification status",
			Error:   err.Error(),
		})
		return
	}

	logger.Log.Info("Notification status updated successfully",
		zap.String("notification_id", notificationID),
		zap.String("status", req.Status),
		zap.String("request_id", requestID.(string)),
	)

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: fmt.Sprintf("Status updated for notification %s: %s", notificationID, req.Status),
		Data: gin.H{
			"notification_id": notificationID,
			"status":          req.Status,
			"updated_at":      time.Now(),
		},
	})
}
