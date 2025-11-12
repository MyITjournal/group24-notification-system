package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupUserTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Add request_id middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("request_id", "test-request-id")
		c.Next()
	})
	return router
}

func TestNewUserHandler(t *testing.T) {
	handler := NewUserHandler()

	assert.NotNil(t, handler)
}

func TestUserHandler_Create_Success(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	userRequest := models.UserCreationRequest{
		Name:      "John Doe",
		Email:     "john@example.com",
		PushToken: "push-token-123",
		Password:  "securepassword",
		Preferences: models.UserPreferences{
			Email: true,
			Push:  false,
		},
	}

	// Create request
	body, _ := json.Marshal(userRequest)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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
	assert.Contains(t, response.Message, "User registration successful")
	assert.Contains(t, response.Message, "User ID:")
	assert.Contains(t, response.Message, "Handed off to User Service")

	// Verify response time header exists
	assert.NotEmpty(t, w.Header().Get("X-Response-Time"))

	// Verify data structure
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "John Doe", data["name"])
	assert.Equal(t, "john@example.com", data["email"])
	assert.Equal(t, "push-token-123", data["push_token"])
	assert.Equal(t, "securepassword", data["password"])

	// Verify preferences
	preferences, ok := data["preferences"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, true, preferences["email_enabled"])
	assert.Equal(t, false, preferences["push_enabled"])
}

func TestUserHandler_Create_InvalidPayload(t *testing.T) {
	tests := []struct {
		name        string
		payload     string
		description string
	}{
		{
			name:        "malformed JSON",
			payload:     `{"name": "John", "email": invalid}`,
			description: "Invalid JSON syntax",
		},
		{
			name:        "missing required fields",
			payload:     `{"name": "John"}`,
			description: "Missing required email field",
		},
		{
			name:        "empty payload",
			payload:     `{}`,
			description: "Empty request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewUserHandler()
			router := setupUserTestRouter()
			router.POST("/users", handler.Create)

			// Create request
			req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.payload))
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
			assert.Equal(t, "Invalid request payload for user creation", response.Message)
			assert.NotEmpty(t, response.Error)
		})
	}
}

func TestUserHandler_Create_ValidatesPreferences(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	userRequest := models.UserCreationRequest{
		Name:     "Jane Smith",
		Email:    "jane@example.com",
		Password: "password123",
		Preferences: models.UserPreferences{
			Email: false,
			Push:  true,
		},
	}

	// Create request
	body, _ := json.Marshal(userRequest)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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

	// Verify preferences are correctly returned
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)

	preferences, ok := data["preferences"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, false, preferences["email_enabled"])
	assert.Equal(t, true, preferences["push_enabled"])
}

func TestUserHandler_Create_OptionalPushToken(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	// Test with empty push token
	userRequest := models.UserCreationRequest{
		Name:      "Bob Jones",
		Email:     "bob@example.com",
		Password:  "password123",
		PushToken: "", // Empty push token
		Preferences: models.UserPreferences{
			Email: true,
			Push:  false,
		},
	}

	// Create request
	body, _ := json.Marshal(userRequest)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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

	// Verify empty push token is accepted
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "", data["push_token"])
}

func TestUserHandler_Create_GeneratesUniqueUserID(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	userRequest := models.UserCreationRequest{
		Name:     "Alice Wonder",
		Email:    "alice@example.com",
		Password: "password123",
		Preferences: models.UserPreferences{
			Email: true,
			Push:  true,
		},
	}

	// Create multiple requests
	userIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(userRequest)
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Extract user ID from message
		assert.Contains(t, response.Message, "User ID:")
		userIDs[i] = response.Message
	}

	// Verify all user IDs are unique
	assert.NotEqual(t, userIDs[0], userIDs[1])
	assert.NotEqual(t, userIDs[1], userIDs[2])
	assert.NotEqual(t, userIDs[0], userIDs[2])
}

func TestUserHandler_Create_AllFieldsPresent(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	userRequest := models.UserCreationRequest{
		Name:      "Complete User",
		Email:     "complete@example.com",
		PushToken: "complete-push-token",
		Password:  "completepassword",
		Preferences: models.UserPreferences{
			Email: true,
			Push:  true,
		},
	}

	// Create request
	body, _ := json.Marshal(userRequest)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
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

	// Verify all fields are present in response
	data, ok := response.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, data["name"])
	assert.NotNil(t, data["email"])
	assert.NotNil(t, data["push_token"])
	assert.NotNil(t, data["password"])
	assert.NotNil(t, data["preferences"])

	preferences, ok := data["preferences"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, preferences["email_enabled"])
	assert.NotNil(t, preferences["push_enabled"])
}

func TestUserHandler_Create_ResponseTimeHeader(t *testing.T) {
	handler := NewUserHandler()
	router := setupUserTestRouter()
	router.POST("/users", handler.Create)

	userRequest := models.UserCreationRequest{
		Name:     "Timing Test",
		Email:    "timing@example.com",
		Password: "password123",
		Preferences: models.UserPreferences{
			Email: true,
			Push:  false,
		},
	}

	// Create request
	body, _ := json.Marshal(userRequest)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify response time header exists and has valid format
	responseTime := w.Header().Get("X-Response-Time")
	assert.NotEmpty(t, responseTime)
	// Should contain time unit (ns, µs, ms, s)
	assert.Regexp(t, `\d+(\.\d+)?(ns|µs|ms|s)`, responseTime)
}
