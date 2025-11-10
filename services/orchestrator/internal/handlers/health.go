package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	// TODO: Add actual dependency checks
	checks := gin.H{
		"user_service":     "ok",
		"template_service": "ok",
		"kafka":            "ok",
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
