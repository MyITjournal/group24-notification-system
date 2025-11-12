package clients

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateClient(t *testing.T) {
	cfg := TemplateClientConfig{
		BaseURL:               "http://localhost:8080",
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      3,
		RetryInitialDelay:     100 * time.Millisecond,
		RetryMaxDelay:         5 * time.Second,
	}

	client := NewTemplateClient(cfg)
	assert.NotNil(t, client)
}

func TestNewTemplateClient_DefaultTimeout(t *testing.T) {
	cfg := TemplateClientConfig{
		BaseURL: "http://localhost:8080",
		// Timeout is 0, should default to 10 seconds
	}

	client := NewTemplateClient(cfg)
	assert.NotNil(t, client)
}

func TestTemplateClient_GetTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/templates/welcome_email", r.URL.Path)
		assert.Equal(t, "en", r.URL.Query().Get("language"))

		template := models.Template{
			TemplateID: "welcome_email",
			Name:       "Welcome Email",
			Version:    "1.0",
			Language:   "en",
			Type:       "email",
			Subject:    "Welcome",
			Body: models.TemplateBody{
				HTML: "<h1>Welcome</h1>",
				Text: "Welcome",
			},
			Variables: []models.TemplateVariable{
				{Name: "name", Type: "string", Required: true},
			},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: "admin",
				Tags:      []string{"welcome"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(template)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	template, err := client.GetTemplate("welcome_email", "en")

	assert.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "welcome_email", template.TemplateID)
	assert.Equal(t, "Welcome Email", template.Name)
}

func TestTemplateClient_GetTemplate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "template not found"}`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	template, err := client.GetTemplate("nonexistent", "en")

	assert.Error(t, err)
	assert.Nil(t, template)
	assert.Contains(t, err.Error(), "non-retryable status 404")
}

func TestTemplateClient_GetTemplate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	template, err := client.GetTemplate("welcome_email", "en")

	assert.Error(t, err)
	assert.Nil(t, template)
	assert.Contains(t, err.Error(), "retryable status 500")
}

func TestTemplateClient_GetTemplate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	template, err := client.GetTemplate("welcome_email", "en")

	assert.Error(t, err)
	assert.Nil(t, template)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestTemplateClient_RenderTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/api/v1/templates/welcome_email/render", r.URL.Path)

		var renderReq models.RenderRequest
		err := json.NewDecoder(r.Body).Decode(&renderReq)
		require.NoError(t, err)
		assert.Equal(t, "en", renderReq.Language)
		assert.Equal(t, "John", renderReq.Variables["name"])

		response := models.RenderResponse{
			TemplateID: "welcome_email",
			Language:   "en",
			Version:    "1.0",
			Rendered: models.RenderedContent{
				Subject: "Welcome John",
				Body: models.TemplateBody{
					HTML: "<h1>Welcome John</h1>",
					Text: "Welcome John",
				},
			},
			RenderedAt:    time.Now(),
			VariablesUsed: []string{"name"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Equal(t, "welcome_email", rendered.TemplateID)
	assert.Equal(t, "Welcome John", rendered.Rendered.Subject)
	assert.Equal(t, "<h1>Welcome John</h1>", rendered.Rendered.Body.HTML)
}

func TestTemplateClient_RenderTemplate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "template not found"}`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("nonexistent", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "non-retryable status 404")
}

func TestTemplateClient_RenderTemplate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "retryable status 500")
}

func TestTemplateClient_RenderTemplate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "failed to unmarshal response")
}

func TestTemplateClient_RenderTemplate_EmptyVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var renderReq models.RenderRequest
		json.NewDecoder(r.Body).Decode(&renderReq)
		assert.Empty(t, renderReq.Variables)

		response := models.RenderResponse{
			TemplateID: "welcome_email",
			Language:   "en",
			Version:    "1.0",
			Rendered: models.RenderedContent{
				Subject: "Welcome",
				Body: models.TemplateBody{
					HTML: "<h1>Welcome</h1>",
					Text: "Welcome",
				},
			},
			RenderedAt:    time.Now(),
			VariablesUsed: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	rendered, err := client.RenderTemplate("welcome_email", "en", nil)

	assert.NoError(t, err)
	assert.NotNil(t, rendered)
}

func TestTemplateClient_RenderTemplate_MultipleVariables(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var renderReq models.RenderRequest
		json.NewDecoder(r.Body).Decode(&renderReq)

		assert.Equal(t, "John", renderReq.Variables["name"])
		assert.Equal(t, "Doe", renderReq.Variables["surname"])
		assert.Equal(t, 30, int(renderReq.Variables["age"].(float64)))

		response := models.RenderResponse{
			TemplateID: "welcome_email",
			Language:   "en",
			Version:    "1.0",
			Rendered: models.RenderedContent{
				Subject: "Welcome John Doe",
				Body: models.TemplateBody{
					HTML: "<h1>Welcome John Doe, age 30</h1>",
					Text: "Welcome John Doe, age 30",
				},
			},
			RenderedAt:    time.Now(),
			VariablesUsed: []string{"name", "surname", "age"},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name":    "John",
		"surname": "Doe",
		"age":     30,
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Equal(t, 3, len(rendered.VariablesUsed))
}

func TestTemplateClient_RenderTemplate_WithRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		response := models.RenderResponse{
			TemplateID: "welcome_email",
			Language:   "en",
			Version:    "1.0",
			Rendered: models.RenderedContent{
				Subject: "Welcome",
				Body: models.TemplateBody{
					HTML: "<h1>Welcome</h1>",
					Text: "Welcome",
				},
			},
			RenderedAt:    time.Now(),
			VariablesUsed: []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      2,
		RetryInitialDelay:     10 * time.Millisecond,
		RetryMaxDelay:         100 * time.Millisecond,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.NoError(t, err)
	assert.NotNil(t, rendered)
	assert.Equal(t, 2, attemptCount)
}

func TestTemplateClient_RenderTemplate_NetworkError(t *testing.T) {
	cfg := TemplateClientConfig{
		BaseURL:               "http://invalid-host:9999",
		Timeout:               1 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "template render request failed")
}

func TestTemplateClient_RenderTemplate_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               100 * time.Millisecond,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
}

func TestTemplateClient_RenderTemplate_TooManyRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": "rate limit exceeded"}`))
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "retryable status 429")
}

func TestTemplateClient_GetTemplate_DifferentLanguages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		language := r.URL.Query().Get("language")

		template := models.Template{
			TemplateID: "welcome_email",
			Name:       "Welcome Email",
			Version:    "1.0",
			Language:   language,
			Type:       "email",
			Subject:    "Welcome",
			Body: models.TemplateBody{
				HTML: "<h1>Welcome</h1>",
				Text: "Welcome",
			},
			Variables: []models.TemplateVariable{},
			Meta: models.TemplateMetadata{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				CreatedBy: "admin",
				Tags:      []string{},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(template)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           5,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           3,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)

	// Test different languages
	languages := []string{"en", "fr", "es", "de"}
	for _, lang := range languages {
		template, err := client.GetTemplate("welcome_email", lang)
		assert.NoError(t, err)
		assert.NotNil(t, template)
		assert.Equal(t, lang, template.Language)
	}
}

func TestTemplateClient_RenderTemplate_CircuitBreakerOpen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cfg := TemplateClientConfig{
		BaseURL:               server.URL,
		Timeout:               5 * time.Second,
		MaxFailures:           2,
		CircuitBreakerTimeout: 60 * time.Second,
		HalfOpenMax:           1,
		RetryMaxAttempts:      0,
	}

	client := NewTemplateClient(cfg)
	variables := map[string]interface{}{
		"name": "John",
	}

	// Trigger enough failures to open circuit breaker
	for i := 0; i < 3; i++ {
		_, _ = client.RenderTemplate("welcome_email", "en", variables)
	}

	// Now circuit should be open
	rendered, err := client.RenderTemplate("welcome_email", "en", variables)

	assert.Error(t, err)
	assert.Nil(t, rendered)
	assert.Contains(t, err.Error(), "temporarily unavailable")
}
