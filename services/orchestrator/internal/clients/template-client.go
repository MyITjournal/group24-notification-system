package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	circuitbreaker "github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/circuit-breaker"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"go.uber.org/zap"
)

type templateClient struct {
	baseURL        string
	httpClient     *http.Client
	circuitBreaker *circuitbreaker.CircuitBreaker
}

type TemplateClientConfig struct {
	BaseURL               string
	Timeout               time.Duration
	MaxFailures           uint32
	CircuitBreakerTimeout time.Duration
	HalfOpenMax           uint32
}

func NewTemplateClient(cfg TemplateClientConfig) TemplateClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &templateClient{
		baseURL: cfg.BaseURL,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		circuitBreaker: circuitbreaker.New(circuitbreaker.Config{
			Name:        "template-service",
			MaxFailures: cfg.MaxFailures,
			Timeout:     cfg.CircuitBreakerTimeout,
			HalfOpenMax: cfg.HalfOpenMax,
		}),
	}
}

func (c *templateClient) GetTemplate(templateID, language string) (*models.Template, error) {
	var template *models.Template

	err := c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/v1/templates/%s?language=%s", c.baseURL, templateID, language)

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		logger.Log.Debug("Calling template service",
			zap.String("url", url),
			zap.String("template_id", templateID),
			zap.String("language", language),
		)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			logger.Log.Error("Template service request failed",
				zap.String("template_id", templateID),
				zap.Error(err),
			)
			return fmt.Errorf("template service request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			logger.Log.Error("Template service returned non-200 status",
				zap.Int("status_code", resp.StatusCode),
				zap.String("template_id", templateID),
				zap.String("response_body", string(body)),
			)
			return fmt.Errorf("template service returned status %d: %s", resp.StatusCode, string(body))
		}

		var result models.Template
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		template = &result
		return nil
	})

	if err != nil {
		if err == circuitbreaker.ErrCircuitOpen {
			logger.Log.Warn("Template service circuit breaker is open",
				zap.String("template_id", templateID),
			)
			return nil, fmt.Errorf("template service is temporarily unavailable: %w", err)
		}
		if err == circuitbreaker.ErrTooManyRequests {
			logger.Log.Warn("Template service circuit breaker: too many requests in half-open state",
				zap.String("template_id", templateID),
			)
			return nil, fmt.Errorf("template service is recovering, please retry: %w", err)
		}
		return nil, err
	}

	return template, nil
}

func (c *templateClient) RenderTemplate(templateID, language string, variables map[string]interface{}) (*models.RenderResponse, error) {
	var rendered *models.RenderResponse

	err := c.circuitBreaker.Execute(func() error {
		url := fmt.Sprintf("%s/api/v1/templates/%s/render", c.baseURL, templateID)

		renderReq := models.RenderRequest{
			Language:    language,
			Version:     "latest",
			Variables:   variables,
			PreviewMode: false,
		}

		reqBody, err := json.Marshal(renderReq)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBody))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")

		logger.Log.Debug("Rendering template",
			zap.String("url", url),
			zap.String("template_id", templateID),
			zap.String("language", language),
		)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			logger.Log.Error("Template render request failed",
				zap.String("template_id", templateID),
				zap.Error(err),
			)
			return fmt.Errorf("template render request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			logger.Log.Error("Template render returned non-200 status",
				zap.Int("status_code", resp.StatusCode),
				zap.String("template_id", templateID),
				zap.String("response_body", string(body)),
			)
			return fmt.Errorf("template render returned status %d: %s", resp.StatusCode, string(body))
		}

		var result models.RenderResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		rendered = &result
		return nil
	})

	if err != nil {
		if err == circuitbreaker.ErrCircuitOpen {
			logger.Log.Warn("Template service circuit breaker is open",
				zap.String("template_id", templateID),
			)
			return nil, fmt.Errorf("template service is temporarily unavailable: %w", err)
		}
		if err == circuitbreaker.ErrTooManyRequests {
			logger.Log.Warn("Template service circuit breaker: too many requests in half-open state",
				zap.String("template_id", templateID),
			)
			return nil, fmt.Errorf("template service is recovering, please retry: %w", err)
		}
		return nil, err
	}

	return rendered, nil
}
