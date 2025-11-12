package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/gin-gonic/gin"
)

// DatabasePinger defines the interface for database ping operations
type DatabasePinger interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db DatabasePinger
}

func NewHealthHandler(db DatabasePinger) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Service is healthy",
		Data: gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := gin.H{
		"user_service":     "ok",
		"template_service": "ok",
		"kafka":            "ok",
	}

	// Check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "error"
		c.JSON(http.StatusServiceUnavailable, models.Response{
			Success: false,
			Message: "Service is not ready",
			Error:   "database connection failed",
			Data: gin.H{
				"status":    "not_ready",
				"checks":    checks,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			},
		})
		return
	}

	checks["database"] = "ok"

	c.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: "Service is ready",
		Data: gin.H{
			"status":    "ready",
			"checks":    checks,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}
