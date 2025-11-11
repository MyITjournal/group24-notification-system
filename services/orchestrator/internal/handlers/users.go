package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// UserHandler handles user-related endpoints like creation and preference updates.
// This handler is responsible for routing and validation. The actual
// interaction with the UserService (creation, storage) will be delegated to the services layer later.
type UserHandler struct {
	// No dependencies needed yet, as it only handles routing/validation before service layer integration.
}

// NewUserHandler initializes a new UserHandler.
// This is the function called by server.go.
func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Create handles POST /api/v1/users for user creation/registration.
func (h *UserHandler) Create(c *gin.Context) {
	requestID, _ := c.Get("request_id")
	startTime := time.Now()

	var req models.UserCreationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Log.Error("Invalid user creation payload",
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

	// Simulation of Service Interaction
	// TODO: This would call: newUserID, err := h.userService.CreateUser(&req) after user service is done
	newUserID := uuid.New().String()

	logger.Log.Info("User creation request received and validated",
		zap.String("request_id", requestID.(string)),
		zap.String("user_email", req.Email),
		zap.Bool("email_pref", req.Preferences.Email),
		zap.Bool("push_pref", req.Preferences.Push),
	)

	// Simulate successful creation and handoff to the User Service.
	duration := time.Since(startTime)
	c.Header("X-Response-Time", duration.String())

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": fmt.Sprintf("User registration successful. User ID: %s. Handed off to User Service.", newUserID),
		"data": gin.H{
			"name":        req.Name,
			"email":       req.Email,
			"push_token":  req.PushToken,
			"preferences": req.Preferences,
			"password":    req.Password,
		},
	})
}
