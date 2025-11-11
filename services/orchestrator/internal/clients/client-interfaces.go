package clients

import (
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type TemplateClient interface {
	GetTemplate(templateID, language string) (*models.Template, error)
	RenderTemplate(templateID, language string, variables map[string]interface{}) (*models.RenderResponse, error)
}

type UserClient interface {
	GetPreferences(userID string) (*models.UserPreferences, error)
}
