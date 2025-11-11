package mocks

import (
	"fmt"
	"strings"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type TemplateServiceMock struct{}

func NewTemplateServiceMock() *TemplateServiceMock {
	return &TemplateServiceMock{}
}

func (m *TemplateServiceMock) GetTemplate(templateID, language string) (*models.Template, error) {
	templates := map[string]*models.Template{
		"welcome_email": {
			TemplateID: "welcome_email",
			Name:       "Welcome Email",
			Version:    "2.3.0",
			Language:   language,
			Type:       "email",
			Subject:    "Welcome to {{app_name}}, {{user_name}}!",
			Body: models.TemplateBody{
				HTML: "<html><body><h1>Welcome {{user_name}}!</h1><p>Thank you for joining {{app_name}}...</p></body></html>",
				Text: "Welcome {{user_name}}!\n\nThank you for joining {{app_name}}...",
			},
			Variables: []models.TemplateVariable{
				{Name: "user_name", Type: "string", Required: true, Description: "User's display name"},
				{Name: "app_name", Type: "string", Required: true, Description: "Application name"},
			},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now().Add(-90 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-5 * 24 * time.Hour),
				CreatedBy: "admin@example.com",
				Tags:      []string{"onboarding", "transactional"},
			},
		},
		"password_reset": {
			TemplateID: "password_reset",
			Name:       "Password Reset",
			Version:    "1.5.2",
			Language:   language,
			Type:       "email",
			Subject:    "Reset your password for {{app_name}}",
			Body: models.TemplateBody{
				HTML: "<html><body><h1>Password Reset Request</h1><p>Click here to reset: {{reset_url}}</p></body></html>",
				Text: "Password Reset Request\n\nClick here to reset: {{reset_url}}",
			},
			Variables: []models.TemplateVariable{
				{Name: "app_name", Type: "string", Required: true, Description: "Application name"},
				{Name: "reset_url", Type: "string", Required: true, Description: "Password reset URL"},
			},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now().Add(-180 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * 24 * time.Hour),
				CreatedBy: "admin@example.com",
				Tags:      []string{"security", "transactional"},
			},
		},
	}

	if template, exists := templates[templateID]; exists {
		return template, nil
	}

	return nil, fmt.Errorf("template not found: %s", templateID)
}

func (m *TemplateServiceMock) RenderTemplate(templateID, language string, variables map[string]interface{}) (*models.RenderResponse, error) {
	template, err := m.GetTemplate(templateID, language)
	if err != nil {
		return nil, err
	}

	// Simple variable substitution
	renderedSubject := template.Subject // Get subject to render variables
	renderedHTML := template.Body.HTML
	renderedText := template.Body.Text

	variablesUsed := []string{}
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)

		// Substitute in Subject
		if strings.Contains(renderedSubject, placeholder) {
			renderedSubject = strings.ReplaceAll(renderedSubject, placeholder, valueStr)
			variablesUsed = append(variablesUsed, key)
		}
		// Substitute in HTML Body
		if strings.Contains(renderedHTML, placeholder) {
			renderedHTML = strings.ReplaceAll(renderedHTML, placeholder, valueStr)
			if !contains(variablesUsed, key) {
				variablesUsed = append(variablesUsed, key)
			}
		}
		// Substitute in Text Body
		if strings.Contains(renderedText, placeholder) {
			renderedText = strings.ReplaceAll(renderedText, placeholder, valueStr)
			if !contains(variablesUsed, key) {
				variablesUsed = append(variablesUsed, key)
			}
		}
	}

	return &models.RenderResponse{
		TemplateID: templateID,
		Language:   language,
		Version:    template.Version,
		Rendered: models.RenderedContent{
			Subject: renderedSubject,
			Body: models.TemplateBody{
				HTML: renderedHTML,
				Text: renderedText,
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: variablesUsed,
	}, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
