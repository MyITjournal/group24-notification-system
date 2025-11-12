package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB is a mock implementation of the database
type MockDB struct {
	mock.Mock
}

func (m *MockDB) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupHealthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewHealthHandler(t *testing.T) {
	mockDB := new(MockDB)
	handler := NewHealthHandler(mockDB)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDB, handler.db)
}

func TestHealthHandler_Live(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "successful health check",
			expectedStatus: http.StatusOK,
			expectedMsg:    "Service is healthy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			handler := NewHealthHandler(mockDB)
			router := setupHealthTestRouter()
			router.GET("/health/live", handler.Live)

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response models.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.True(t, response.Success)
			assert.Equal(t, tt.expectedMsg, response.Message)

			// Verify data structure
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, "ok", data["status"])
			assert.NotEmpty(t, data["timestamp"])

			// Verify timestamp format
			timestamp, ok := data["timestamp"].(string)
			assert.True(t, ok)
			_, err = time.Parse(time.RFC3339, timestamp)
			assert.NoError(t, err)
		})
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	tests := []struct {
		name            string
		setupMock       func(*MockDB)
		expectedStatus  int
		expectedMsg     string
		expectedSuccess bool
		checkDatabase   string
	}{
		{
			name: "all services ready - database connected",
			setupMock: func(m *MockDB) {
				m.On("Ping", mock.AnythingOfType("*context.timerCtx")).Return(nil)
			},
			expectedStatus:  http.StatusOK,
			expectedMsg:     "Service is ready",
			expectedSuccess: true,
			checkDatabase:   "ok",
		},
		{
			name: "database connection failed",
			setupMock: func(m *MockDB) {
				m.On("Ping", mock.AnythingOfType("*context.timerCtx")).
					Return(errors.New("connection refused"))
			},
			expectedStatus:  http.StatusServiceUnavailable,
			expectedMsg:     "Service is not ready",
			expectedSuccess: false,
			checkDatabase:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := new(MockDB)
			tt.setupMock(mockDB)
			handler := NewHealthHandler(mockDB)
			router := setupHealthTestRouter()
			router.GET("/health/ready", handler.Ready)

			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
			w := httptest.NewRecorder()

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response models.Response
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSuccess, response.Success)
			assert.Equal(t, tt.expectedMsg, response.Message)

			// Verify data structure
			data, ok := response.Data.(map[string]interface{})
			assert.True(t, ok)

			checks, ok := data["checks"].(map[string]interface{})
			assert.True(t, ok)

			// Verify all checks exist
			assert.Equal(t, "ok", checks["user_service"])
			assert.Equal(t, "ok", checks["template_service"])
			assert.Equal(t, "ok", checks["kafka"])
			assert.Equal(t, tt.checkDatabase, checks["database"])

			// Verify timestamp
			timestamp, ok := data["timestamp"].(string)
			assert.True(t, ok)
			_, err = time.Parse(time.RFC3339, timestamp)
			assert.NoError(t, err)

			// Verify status field
			if tt.expectedSuccess {
				assert.Equal(t, "ready", data["status"])
			} else {
				assert.Equal(t, "not_ready", data["status"])
				assert.Equal(t, "database connection failed", response.Error)
			}

			// Verify mock expectations
			mockDB.AssertExpectations(t)
		})
	}
}

func TestHealthHandler_Ready_ContextTimeout(t *testing.T) {
	// This test verifies that the context timeout is properly set
	mockDB := new(MockDB)
	mockDB.On("Ping", mock.AnythingOfType("*context.timerCtx")).
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			deadline, ok := ctx.Deadline()
			assert.True(t, ok)
			// Verify timeout is approximately 2 seconds
			assert.WithinDuration(t, time.Now().Add(2*time.Second), deadline, 100*time.Millisecond)
		}).
		Return(nil)

	handler := NewHealthHandler(mockDB)
	router := setupHealthTestRouter()
	router.GET("/health/ready", handler.Ready)

	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockDB.AssertExpectations(t)
}
