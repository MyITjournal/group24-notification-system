package mocks

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type TemplateServiceMock struct {
	// Configurable behavior
	simulateDelay    bool
	simulateErrors   bool
	errorRate        float64
	delayMin         time.Duration
	delayMax         time.Duration
	requestCount     int
	failureThreshold int
}

func NewTemplateServiceMock() *TemplateServiceMock {
	// Read configuration from environment variables
	simulateDelay := os.Getenv("MOCK_SIMULATE_DELAY") == "true"
	simulateErrors := os.Getenv("MOCK_SIMULATE_ERRORS") == "true"

	var errorRate float64 = 0.1
	if rate := os.Getenv("MOCK_ERROR_RATE"); rate != "" {
		if _, err := fmt.Sscanf(rate, "%f", &errorRate); err != nil {
			errorRate = 0.1
		}
	}

	delayMin := 50 * time.Millisecond
	delayMax := 200 * time.Millisecond
	if min := os.Getenv("MOCK_DELAY_MIN"); min != "" {
		if d, err := time.ParseDuration(min); err == nil {
			delayMin = d
		}
	}
	if max := os.Getenv("MOCK_DELAY_MAX"); max != "" {
		if d, err := time.ParseDuration(max); err == nil {
			delayMax = d
		}
	}

	return &TemplateServiceMock{
		simulateDelay:    simulateDelay,
		simulateErrors:   simulateErrors,
		errorRate:        errorRate,
		delayMin:         delayMin,
		delayMax:         delayMax,
		requestCount:     0,
		failureThreshold: 0,
	}
}

// NewTemplateServiceMockWithConfig creates a mock with custom configuration
func NewTemplateServiceMockWithConfig(cfg MockBehaviorConfig) *TemplateServiceMock {
	return &TemplateServiceMock{
		simulateDelay:    cfg.SimulateDelay,
		simulateErrors:   cfg.SimulateErrors,
		errorRate:        cfg.ErrorRate,
		delayMin:         cfg.DelayMin,
		delayMax:         cfg.DelayMax,
		requestCount:     0,
		failureThreshold: cfg.FailureThreshold,
	}
}

func (m *TemplateServiceMock) GetTemplate(templateID, language string) (*models.Template, error) {
	m.requestCount++

	// Simulate network delay
	if m.simulateDelay {
		delay := m.delayMin + time.Duration(rand.Int63n(int64(m.delayMax-m.delayMin)))
		time.Sleep(delay)
	}

	// Simulate failure threshold
	if m.failureThreshold > 0 && m.requestCount > m.failureThreshold {
		return nil, fmt.Errorf("template service unavailable: service overloaded")
	}

	// Simulate random errors
	if m.simulateErrors && rand.Float64() < m.errorRate {
		return m.simulateError(templateID)
	}

	// Simulate template not found
	if strings.HasPrefix(templateID, "notfound_") || templateID == "nonexistent_template" {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Simulate timeout
	if strings.HasPrefix(templateID, "timeout_") {
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("request timeout: template service did not respond")
	}

	// Simulate service unavailable
	if strings.HasPrefix(templateID, "unavailable_") {
		return nil, fmt.Errorf("template service unavailable: service temporarily down")
	}

	// Simulate invalid template (missing required fields)
	if strings.HasPrefix(templateID, "invalid_") {
		return nil, fmt.Errorf("invalid template format: %s", templateID)
	}

	// Simulate version not found
	if strings.Contains(templateID, "_v999") {
		return nil, fmt.Errorf("template version not found: %s", templateID)
	}

	// Get template from predefined set
	template := m.getTemplate(templateID, language)
	if template == nil {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

func (m *TemplateServiceMock) RenderTemplate(templateID, language string, variables map[string]interface{}) (*models.RenderResponse, error) {
	m.requestCount++

	// Simulate network delay
	if m.simulateDelay {
		delay := m.delayMin + time.Duration(rand.Int63n(int64(m.delayMax-m.delayMin)))
		time.Sleep(delay)
	}

	// Simulate failure threshold
	if m.failureThreshold > 0 && m.requestCount > m.failureThreshold {
		return nil, fmt.Errorf("template service unavailable: service overloaded")
	}

	// Simulate random errors
	if m.simulateErrors && rand.Float64() < m.errorRate {
		return m.simulateRenderError(templateID)
	}

	// Get template first
	template, err := m.GetTemplate(templateID, language)
	if err != nil {
		return nil, err
	}

	// Simulate missing required variables
	if strings.HasPrefix(templateID, "missing_vars_") {
		requiredVars := []string{}
		for _, v := range template.Variables {
			if v.Required {
				requiredVars = append(requiredVars, v.Name)
			}
		}
		missing := []string{}
		for _, req := range requiredVars {
			if _, ok := variables[req]; !ok {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("missing required variables: %v", missing)
		}
	}

	// Simulate invalid variable types
	if strings.HasPrefix(templateID, "invalid_vars_") {
		return nil, fmt.Errorf("invalid variable type for template: %s", templateID)
	}

	// Simulate rendering timeout
	if strings.HasPrefix(templateID, "render_timeout_") {
		time.Sleep(5 * time.Second)
		return nil, fmt.Errorf("template rendering timeout: %s", templateID)
	}

	// Perform variable substitution
	renderedSubject := template.Subject
	renderedHTML := template.Body.HTML
	renderedText := template.Body.Text

	variablesUsed := []string{}
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		valueStr := fmt.Sprintf("%v", value)

		// Substitute in Subject
		if strings.Contains(renderedSubject, placeholder) {
			renderedSubject = strings.ReplaceAll(renderedSubject, placeholder, valueStr)
			if !contains(variablesUsed, key) {
				variablesUsed = append(variablesUsed, key)
			}
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

	// Check for missing required variables
	for _, v := range template.Variables {
		if v.Required && !contains(variablesUsed, v.Name) {
			return nil, fmt.Errorf("missing required variable: %s", v.Name)
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

func (m *TemplateServiceMock) getTemplate(templateID, language string) *models.Template {
	// Extract base template ID (remove prefixes used for error simulation)
	baseID := templateID
	if idx := strings.Index(templateID, "_"); idx > 0 {
		// Keep the base ID for lookup
		baseID = templateID
	}

	templates := map[string]*models.Template{
		"welcome_email": {
			TemplateID: "welcome_email",
			Name:       "Welcome Email",
			Version:    "2.3.0",
			Language:   language,
			Type:       "email",
			Subject:    "Welcome to {{app_name}}, {{user_name}}!",
			Body: models.TemplateBody{
				HTML: "<html><body><h1>Welcome {{user_name}}!</h1><p>Thank you for joining {{app_name}}. We're excited to have you on board!</p><p>Get started by exploring our features.</p></body></html>",
				Text: "Welcome {{user_name}}!\n\nThank you for joining {{app_name}}. We're excited to have you on board!\n\nGet started by exploring our features.",
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
				HTML: "<html><body><h1>Password Reset Request</h1><p>Click the link below to reset your password:</p><p><a href=\"{{reset_url}}\">Reset Password</a></p><p>This link will expire in 1 hour.</p></body></html>",
				Text: "Password Reset Request\n\nClick the link below to reset your password:\n{{reset_url}}\n\nThis link will expire in 1 hour.",
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
		"push_notification": {
			TemplateID: "push_notification",
			Name:       "Push Notification",
			Version:    "1.0.0",
			Language:   language,
			Type:       "push",
			Subject:    "{{title}}",
			Body: models.TemplateBody{
				HTML: "",
				Text: "{{message}}\n\n{{action_text}}: {{action_url}}",
			},
			Variables: []models.TemplateVariable{
				{Name: "title", Type: "string", Required: true, Description: "Notification title"},
				{Name: "message", Type: "string", Required: true, Description: "Notification message"},
				{Name: "action_text", Type: "string", Required: false, Description: "Action button text"},
				{Name: "action_url", Type: "string", Required: false, Description: "Action URL"},
			},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now().Add(-60 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-10 * 24 * time.Hour),
				CreatedBy: "admin@example.com",
				Tags:      []string{"push", "notification"},
			},
		},
		"order_confirmation": {
			TemplateID: "order_confirmation",
			Name:       "Order Confirmation",
			Version:    "3.1.0",
			Language:   language,
			Type:       "email",
			Subject:    "Order #{{order_id}} Confirmed",
			Body: models.TemplateBody{
				HTML: "<html><body><h1>Order Confirmed!</h1><p>Hi {{user_name}},</p><p>Your order #{{order_id}} has been confirmed.</p><p>Total: {{order_total}}</p><p>Thank you for your purchase!</p></body></html>",
				Text: "Order Confirmed!\n\nHi {{user_name}},\n\nYour order #{{order_id}} has been confirmed.\n\nTotal: {{order_total}}\n\nThank you for your purchase!",
			},
			Variables: []models.TemplateVariable{
				{Name: "user_name", Type: "string", Required: true, Description: "User's name"},
				{Name: "order_id", Type: "string", Required: true, Description: "Order ID"},
				{Name: "order_total", Type: "string", Required: true, Description: "Order total amount"},
			},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now().Add(-120 * 24 * time.Hour),
				UpdatedAt: time.Now().Add(-7 * 24 * time.Hour),
				CreatedBy: "admin@example.com",
				Tags:      []string{"transactional", "commerce"},
			},
		},
	}

	// Try exact match first
	if template, exists := templates[templateID]; exists {
		return template
	}

	// Try base ID match (for error simulation prefixes)
	for key, template := range templates {
		if strings.HasSuffix(baseID, key) || strings.Contains(baseID, key) {
			// Return a copy with the requested template ID
			result := *template
			result.TemplateID = templateID
			return &result
		}
	}

	return nil
}

func (m *TemplateServiceMock) simulateError(templateID string) (*models.Template, error) {
	errorTypes := []string{
		"internal_server_error",
		"service_unavailable",
		"timeout",
		"database_error",
		"network_error",
	}

	errorType := errorTypes[rand.Intn(len(errorTypes))]

	switch errorType {
	case "internal_server_error":
		return nil, fmt.Errorf("template service internal server error: %s", templateID)
	case "service_unavailable":
		return nil, fmt.Errorf("template service unavailable: %s", templateID)
	case "timeout":
		return nil, fmt.Errorf("template service request timeout: %s", templateID)
	case "database_error":
		return nil, fmt.Errorf("template service database error: %s", templateID)
	case "network_error":
		return nil, fmt.Errorf("template service network error: connection refused")
	default:
		return nil, fmt.Errorf("template service error: %s", templateID)
	}
}

func (m *TemplateServiceMock) simulateRenderError(templateID string) (*models.RenderResponse, error) {
	errorTypes := []string{
		"rendering_failed",
		"invalid_variables",
		"template_corrupted",
		"service_unavailable",
	}

	errorType := errorTypes[rand.Intn(len(errorTypes))]

	switch errorType {
	case "rendering_failed":
		return nil, fmt.Errorf("template rendering failed: %s", templateID)
	case "invalid_variables":
		return nil, fmt.Errorf("invalid variables provided for template: %s", templateID)
	case "template_corrupted":
		return nil, fmt.Errorf("template data corrupted: %s", templateID)
	case "service_unavailable":
		return nil, fmt.Errorf("template service unavailable: %s", templateID)
	default:
		return nil, fmt.Errorf("template rendering error: %s", templateID)
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Reset resets the mock state (useful for testing)
func (m *TemplateServiceMock) Reset() {
	m.requestCount = 0
}

// GetRequestCount returns the number of requests made (for testing)
func (m *TemplateServiceMock) GetRequestCount() int {
	return m.requestCount
}
