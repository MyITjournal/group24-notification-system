package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	// Initialize logger for tests
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
	defer logger.Sync()

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// MockOrchestrationService mocks the orchestration service
type MockOrchestrationService struct {
	mock.Mock
}

func (m *MockOrchestrationService) ProcessNotification(req *models.NotificationRequest) (*models.NotificationResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationResponse), args.Error(1)
}

func (m *MockOrchestrationService) UpdateNotificationStatus(ctx context.Context, notificationID string, status models.NotificationStatus, errorMsg string) error {
	args := m.Called(ctx, notificationID, status, errorMsg)
	return args.Error(0)
}

// MockIdempotencyService mocks the idempotency service
type MockIdempotencyService struct {
	mock.Mock
}

func (m *MockIdempotencyService) GetCachedResponse(ctx context.Context, key string) (*models.NotificationResponse, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationResponse), args.Error(1)
}

func (m *MockIdempotencyService) StoreResponse(ctx context.Context, key string, response *models.NotificationResponse) error {
	args := m.Called(ctx, key, response)
	return args.Error(0)
}

func setupNotificationTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Add request_id middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	return router
}

func TestNewNotificationHandler(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	handler := NewNotificationHandler(mockOrch, mockIdem)

	assert.NotNil(t, handler)
	assert.Equal(t, mockOrch, handler.orchestrationService)
	assert.Equal(t, mockIdem, handler.idempotencyService)
}

func TestNotificationHandler_Create_Success(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	notifRequest := &models.NotificationRequest{
		RequestID:        "req-123",
		UserID:           "user-456",
		NotificationType: models.NotificationEmail,
		TemplateCode:     "test_template",
	}

	notifResponse := &models.NotificationResponse{
		NotificationID: "notif-789",
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}

	// Mock expectations
	mockIdem.On("GetCachedResponse", mock.Anything, "req-123").Return(nil, nil)
	mockOrch.On("ProcessNotification", notifRequest).Return(notifResponse, nil)
	mockIdem.On("StoreResponse", mock.Anything, "req-123", notifResponse).Return(nil)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/notifications", handler.Create)

	// Create request
	body, _ := json.Marshal(notifRequest)
	req, _ := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Notification queued successfully", response.Message)

	// Verify response time header exists
	assert.NotEmpty(t, w.Header().Get("X-Response-Time"))

	mockOrch.AssertExpectations(t)
	mockIdem.AssertExpectations(t)
}

func TestNotificationHandler_Create_CachedResponse(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	notifRequest := &models.NotificationRequest{
		RequestID:        "req-123",
		UserID:           "user-456",
		NotificationType: models.NotificationEmail,
		TemplateCode:     "test_template",
	}

	cachedResponse := &models.NotificationResponse{
		NotificationID: "notif-789",
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}

	// Mock expectations - return cached response
	mockIdem.On("GetCachedResponse", mock.Anything, "req-123").Return(cachedResponse, nil)
	// ProcessNotification should NOT be called
	// StoreResponse should NOT be called

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/notifications", handler.Create)

	// Create request
	body, _ := json.Marshal(notifRequest)
	req, _ := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Notification retrieved from cache", response.Message)

	// Verify idempotent replay header
	assert.Equal(t, "true", w.Header().Get("X-Idempotent-Replay"))
	assert.NotEmpty(t, w.Header().Get("X-Response-Time"))

	mockIdem.AssertExpectations(t)
	mockOrch.AssertNotCalled(t, "ProcessNotification")
}

func TestNotificationHandler_Create_InvalidPayload(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/notifications", handler.Create)

	// Create invalid request
	invalidJSON := []byte(`{"invalid": json}`)
	req, _ := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request payload", response.Message)
	assert.NotEmpty(t, response.Error)
}

func TestNotificationHandler_Create_ProcessingError(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	notifRequest := &models.NotificationRequest{
		RequestID:        "req-123",
		UserID:           "user-456",
		NotificationType: models.NotificationEmail,
		TemplateCode:     "test_template",
	}

	// Mock expectations
	mockIdem.On("GetCachedResponse", mock.Anything, "req-123").Return(nil, nil)
	mockOrch.On("ProcessNotification", notifRequest).Return(nil, errors.New("processing failed"))

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/notifications", handler.Create)

	// Create request
	body, _ := json.Marshal(notifRequest)
	req, _ := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Failed to process notification", response.Message)
	assert.Equal(t, "processing failed", response.Error)

	mockOrch.AssertExpectations(t)
	mockIdem.AssertExpectations(t)
}

func TestNotificationHandler_Create_IdempotencyCheckError(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	notifRequest := &models.NotificationRequest{
		RequestID:        "req-123",
		UserID:           "user-456",
		NotificationType: models.NotificationEmail,
		TemplateCode:     "test_template",
	}

	notifResponse := &models.NotificationResponse{
		NotificationID: "notif-789",
		Status:         models.StatusPending,
		Timestamp:      time.Now(),
	}

	// Mock expectations - idempotency check fails but processing continues
	mockIdem.On("GetCachedResponse", mock.Anything, "req-123").Return(nil, errors.New("redis error"))
	mockOrch.On("ProcessNotification", notifRequest).Return(notifResponse, nil)
	mockIdem.On("StoreResponse", mock.Anything, "req-123", notifResponse).Return(nil)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/notifications", handler.Create)

	// Create request
	body, _ := json.Marshal(notifRequest)
	req, _ := http.NewRequest(http.MethodPost, "/notifications", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert - should succeed despite idempotency check error
	assert.Equal(t, http.StatusCreated, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockOrch.AssertExpectations(t)
	mockIdem.AssertExpectations(t)
}

// TestNotificationHandler_Create_SkippedStatus removed - "skipped" is not a valid NotificationStatus

func TestNotificationHandler_UpdateStatus_Success(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	updateRequest := map[string]interface{}{
		"notification_id": "notif-123",
		"status":          "delivered",
		"timestamp":       time.Now(),
	}

	// Mock expectations
	mockOrch.On("UpdateNotificationStatus", mock.Anything, "notif-123", models.StatusDelivered, "").Return(nil)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/status", handler.UpdateStatus)

	// Create request
	body, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest(http.MethodPost, "/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "Status updated for notification notif-123")

	mockOrch.AssertExpectations(t)
}

func TestNotificationHandler_UpdateStatus_InvalidPayload(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/status", handler.UpdateStatus)

	// Create invalid request (missing required fields)
	invalidRequest := map[string]interface{}{
		"status": "delivered",
	}

	body, _ := json.Marshal(invalidRequest)
	req, _ := http.NewRequest(http.MethodPost, "/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Invalid request payload", response.Message)
}

func TestNotificationHandler_UpdateStatus_InvalidStatus(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/status", handler.UpdateStatus)

	// Create request with invalid status
	invalidRequest := map[string]interface{}{
		"notification_id": "notif-123",
		"status":          "invalid_status",
		"timestamp":       time.Now(),
	}

	body, _ := json.Marshal(invalidRequest)
	req, _ := http.NewRequest(http.MethodPost, "/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestNotificationHandler_UpdateStatus_DatabaseError(t *testing.T) {
	mockOrch := new(MockOrchestrationService)
	mockIdem := new(MockIdempotencyService)

	updateRequest := map[string]interface{}{
		"notification_id": "notif-123",
		"status":          "failed",
		"timestamp":       time.Now(),
		"error":           "delivery failed",
	}

	// Mock expectations - database update fails
	mockOrch.On("UpdateNotificationStatus", mock.Anything, "notif-123", models.StatusFailed, "delivery failed").
		Return(errors.New("database error"))

	handler := NewNotificationHandler(mockOrch, mockIdem)
	router := setupNotificationTestRouter()
	router.POST("/status", handler.UpdateStatus)

	// Create request
	body, _ := json.Marshal(updateRequest)
	req, _ := http.NewRequest(http.MethodPost, "/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Failed to update notification status", response.Message)
	assert.Equal(t, "database error", response.Error)

	mockOrch.AssertExpectations(t)
}
