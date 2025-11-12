package middleware

import (
	"net/http"
	"os"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/gin-gonic/gin"
)

// APIKeyAuth validates API key from header or query parameter
// If no API_KEY environment variable is set, it allows all requests (development mode)
func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		expectedKey := os.Getenv("API_KEY")
		if expectedKey == "" {
			// Development mode - allow all if no key configured
			c.Next()
			return
		}

		if apiKey != expectedKey {
			c.JSON(http.StatusUnauthorized, models.Response{
				Success: false,
				Message: "Unauthorized",
				Error:   "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
