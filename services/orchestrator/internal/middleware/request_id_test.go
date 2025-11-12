package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
	defer logger.Sync()
	code := m.Run()
	os.Exit(code)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func TestRequestID_WithHeader(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "custom-request-id-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "custom-request-id-123", w.Header().Get("X-Request-ID"))
}

func TestRequestID_WithoutHeader(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should generate a new UUID
	responseHeaderID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, responseHeaderID)

	// Verify it's a valid UUID
	_, err := uuid.Parse(responseHeaderID)
	assert.NoError(t, err)
}

func TestRequestID_EmptyHeader(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Should generate a new UUID when header is empty
	responseHeaderID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, responseHeaderID)

	_, err := uuid.Parse(responseHeaderID)
	assert.NoError(t, err)
}

func TestRequestID_ContextValue(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())

	var capturedRequestID string
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		capturedRequestID = requestID.(string)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-id-456")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test-id-456", capturedRequestID)
	assert.Equal(t, "test-id-456", w.Header().Get("X-Request-ID"))
}

func TestRequestID_UniqueGeneration(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Make multiple requests without X-Request-ID header
	requestIDs := make(map[string]bool)
	for i := 0; i < 10; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)

		// Verify each request gets a unique ID
		assert.False(t, requestIDs[requestID], "Request ID should be unique")
		requestIDs[requestID] = true
	}
}

func TestRequestID_MultipleRequests(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("request_id")
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	// Test with different request IDs
	testIDs := []string{"id-1", "id-2", "id-3"}

	for _, testID := range testIDs {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", testID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, testID, w.Header().Get("X-Request-ID"))
	}
}
