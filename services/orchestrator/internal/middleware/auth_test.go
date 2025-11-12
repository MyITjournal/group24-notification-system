package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuth_NoAPIKeyConfigured_AllowsAll(t *testing.T) {
	// Save original value
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	// Remove API_KEY for this test
	os.Unsetenv("API_KEY")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_ValidKeyInHeader(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "test-api-key-123")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test-api-key-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_ValidKeyInQuery(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "test-api-key-123")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test?api_key=test-api-key-123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_InvalidKeyInHeader(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Unauthorized", response.Message)
	assert.Contains(t, response.Error, "Invalid or missing API key")
}

func TestAPIKeyAuth_InvalidKeyInQuery(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test?api_key=wrong-key", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
}

func TestAPIKeyAuth_MissingKey(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "required-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// No API key provided
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response models.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "Invalid or missing API key")
}

func TestAPIKeyAuth_HeaderTakesPrecedence(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test?api_key=wrong-key", nil)
	req.Header.Set("X-API-Key", "correct-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should succeed because header takes precedence
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_EmptyHeaderFallsBackToQuery(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test?api_key=correct-key", nil)
	req.Header.Set("X-API-Key", "") // Empty header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_CaseSensitive(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "Test-Key-123")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with different case
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "test-key-123") // Different case
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyAuth_CorrectCase(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "Test-Key-123")

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "Test-Key-123") // Exact match
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_AbortsOnInvalidKey(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())

	handlerCalled := false
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerCalled, "Handler should not be called when auth fails")
}

func TestAPIKeyAuth_ContinuesOnValidKey(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	os.Setenv("API_KEY", "correct-key")

	router := setupTestRouter()
	router.Use(APIKeyAuth())

	handlerCalled := false
	router.GET("/test", func(c *gin.Context) {
		handlerCalled = true
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "correct-key")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "Handler should be called when auth succeeds")
}

func TestAPIKeyAuth_SpecialCharactersInKey(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	testKey := "key-with-special-chars-!@#$%^&*()"
	os.Setenv("API_KEY", testKey)

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", testKey)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyAuth_LongKey(t *testing.T) {
	originalKey := os.Getenv("API_KEY")
	defer os.Setenv("API_KEY", originalKey)

	// Generate a long key
	longKey := "a"
	for i := 0; i < 1000; i++ {
		longKey += "a"
	}
	os.Setenv("API_KEY", longKey)

	router := setupTestRouter()
	router.Use(APIKeyAuth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", longKey)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
