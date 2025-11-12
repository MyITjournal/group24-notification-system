package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogging_BasicRequest(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID()) // Need request ID for logging
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_WithQueryParams(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test?param1=value1&param2=value2", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_DifferentMethods(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		router.Handle(method, "/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"method": method})
		})

		req, _ := http.NewRequest(method, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestLogging_DifferentStatusCodes(t *testing.T) {
	statusCodes := []int{
		http.StatusOK,
		http.StatusCreated,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		// Create a new router for each status code to avoid route conflicts
		router := setupTestRouter()
		router.Use(RequestID())
		router.Use(Logging())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(statusCode, gin.H{"status": statusCode})
		})

		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, statusCode, w.Code)
	}
}

func TestLogging_WithRequestID(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id-123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_Duration(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	// Duration should be at least 50ms (with some tolerance)
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
}

func TestLogging_ClientIP(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ip": c.ClientIP()})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_EmptyQuery(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogging_ErrorResponse(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestLogging_DifferentPaths(t *testing.T) {
	router := setupTestRouter()
	router.Use(RequestID())
	router.Use(Logging())

	paths := []string{"/api/v1/notifications", "/api/v1/users", "/health", "/test/path/with/multiple/segments"}

	for _, path := range paths {
		router.GET(path, func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"path": c.Request.URL.Path})
		})

		req, _ := http.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}
}
