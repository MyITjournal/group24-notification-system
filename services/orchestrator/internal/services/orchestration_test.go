package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserClient mocks the UserClient interface
type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) GetPreferences(userID string) (*models.UserPreferences, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPreferences), args.Error(1)
}

// MockTemplateClient mocks the TemplateClient interface
type MockTemplateClient struct {
	mock.Mock
}

func (m *MockTemplateClient) GetTemplate(templateID, language string) (*models.Template, error) {
	args := m.Called(templateID, language)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func (m *MockTemplateClient) RenderTemplate(templateID, language string, variables map[string]interface{}) (*models.RenderResponse, error) {
	args := m.Called(templateID, language, variables)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RenderResponse), args.Error(1)
}

// MockKafkaManager mocks the Kafka Manager
type MockKafkaManager struct {
	mock.Mock
}

func (m *MockKafkaManager) PublishEmail(ctx context.Context, notificationID string, payload interface{}) error {
	args := m.Called(ctx, notificationID, payload)
	return args.Error(0)
}

func (m *MockKafkaManager) PublishPush(ctx context.Context, notificationID string, payload interface{}) error {
	args := m.Called(ctx, notificationID, payload)
	return args.Error(0)
}

func (m *MockKafkaManager) PublishByType(ctx context.Context, notificationType, notificationID string, payload interface{}) error {
	args := m.Called(ctx, notificationType, notificationID, payload)
	return args.Error(0)
}

func (m *MockKafkaManager) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockKafkaManager) HealthCheck() error {
	args := m.Called()
	return args.Error(0)
}

// MockNotificationRepository mocks the NotificationRepository interface
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *models.NotificationRecord) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id string) (*models.NotificationRecord, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.NotificationRecord), args.Error(1)
}

func (m *MockNotificationRepository) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error {
	args := m.Called(ctx, id, status, errorMsg)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.NotificationRecord, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.NotificationRecord), args.Error(1)
}

func TestNewOrchestrationService(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	assert.NotNil(t, service)
	assert.Equal(t, mockUserClient, service.userClient)
	assert.Equal(t, mockTemplateClient, service.templateClient)
	assert.Equal(t, mockKafkaManager, service.kafkaManager)
	assert.Equal(t, mockRepo, service.notificationRepo)
}

func TestOrchestrationService_ProcessNotification_Success_Email(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
		Priority:         2,
	}

	userPrefs := &models.UserPreferences{
		Email: true,
		Push:  false,
	}

	rendered := &models.RenderResponse{
		TemplateID: "welcome_email",
		Language:   "en",
		Version:    "1.0",
		Rendered: models.RenderedContent{
			Subject: "Welcome",
			Body: models.TemplateBody{
				HTML: "<h1>Welcome John</h1>",
				Text: "Welcome John",
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: []string{"name"},
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockTemplateClient.On("RenderTemplate", "welcome_email", "en", req.Variables).Return(rendered, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NotificationRecord")).Return(nil)
	mockKafkaManager.On("PublishByType", mock.Anything, "email", mock.AnythingOfType("string"), mock.AnythingOfType("*models.KafkaNotificationPayload")).Return(nil)

	response, err := service.ProcessNotification(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.NotificationID)
	assert.Equal(t, models.StatusPending, response.Status)
	assert.Empty(t, response.Error)

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockKafkaManager.AssertExpectations(t)
}

func TestOrchestrationService_ProcessNotification_Success_Push(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationPush,
		UserID:           "user-456",
		TemplateCode:     "push_notification",
		Variables:        map[string]interface{}{"message": "Hello"},
		Priority:         3,
	}

	userPrefs := &models.UserPreferences{
		Email: false,
		Push:  true,
	}

	rendered := &models.RenderResponse{
		TemplateID: "push_notification",
		Language:   "en",
		Version:    "1.0",
		Rendered: models.RenderedContent{
			Subject: "Push Notification",
			Body: models.TemplateBody{
				HTML: "",
				Text: "Hello",
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: []string{"message"},
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockTemplateClient.On("RenderTemplate", "push_notification", "en", req.Variables).Return(rendered, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NotificationRecord")).Return(nil)
	mockKafkaManager.On("PublishByType", mock.Anything, "push", mock.AnythingOfType("string"), mock.AnythingOfType("*models.KafkaNotificationPayload")).Return(nil)

	response, err := service.ProcessNotification(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.NotificationID)
	assert.Equal(t, models.StatusPending, response.Status)

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockKafkaManager.AssertExpectations(t)
}

func TestOrchestrationService_ProcessNotification_UserPreferencesError(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
	}

	// Mock expectations - user service fails
	mockUserClient.On("GetPreferences", "user-456").Return(nil, errors.New("user service unavailable"))

	response, err := service.ProcessNotification(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to get user preferences")

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertNotCalled(t, "RenderTemplate")
	mockRepo.AssertNotCalled(t, "Create")
	mockKafkaManager.AssertNotCalled(t, "PublishByType")
}

func TestOrchestrationService_ProcessNotification_EmailDisabled(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
	}

	userPrefs := &models.UserPreferences{
		Email: false, // Email disabled
		Push:  true,
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NotificationRecord")).Return(nil)

	response, err := service.ProcessNotification(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusFailed, response.Status)
	assert.Contains(t, response.Error, "email notifications disabled")

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertNotCalled(t, "RenderTemplate")
	mockRepo.AssertExpectations(t)
	mockKafkaManager.AssertNotCalled(t, "PublishByType")
}

func TestOrchestrationService_ProcessNotification_PushDisabled(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationPush,
		UserID:           "user-456",
		TemplateCode:     "push_notification",
		Variables:        map[string]interface{}{"message": "Hello"},
	}

	userPrefs := &models.UserPreferences{
		Email: true,
		Push:  false, // Push disabled
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NotificationRecord")).Return(nil)

	response, err := service.ProcessNotification(req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, models.StatusFailed, response.Status)
	assert.Contains(t, response.Error, "push notifications disabled")

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertNotCalled(t, "RenderTemplate")
	mockRepo.AssertExpectations(t)
	mockKafkaManager.AssertNotCalled(t, "PublishByType")
}

func TestOrchestrationService_ProcessNotification_TemplateRenderError(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
	}

	userPrefs := &models.UserPreferences{
		Email: true,
		Push:  false,
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockTemplateClient.On("RenderTemplate", "welcome_email", "en", req.Variables).Return(nil, errors.New("template not found"))

	response, err := service.ProcessNotification(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to render template")

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "Create")
	mockKafkaManager.AssertNotCalled(t, "PublishByType")
}

func TestOrchestrationService_ProcessNotification_KafkaPublishError(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
	}

	userPrefs := &models.UserPreferences{
		Email: true,
		Push:  false,
	}

	rendered := &models.RenderResponse{
		TemplateID: "welcome_email",
		Language:   "en",
		Version:    "1.0",
		Rendered: models.RenderedContent{
			Subject: "Welcome",
			Body: models.TemplateBody{
				HTML: "<h1>Welcome John</h1>",
				Text: "Welcome John",
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: []string{"name"},
	}

	// Mock expectations
	mockUserClient.On("GetPreferences", "user-456").Return(userPrefs, nil)
	mockTemplateClient.On("RenderTemplate", "welcome_email", "en", req.Variables).Return(rendered, nil)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.NotificationRecord")).Return(nil)
	mockKafkaManager.On("PublishByType", mock.Anything, "email", mock.AnythingOfType("string"), mock.AnythingOfType("*models.KafkaNotificationPayload")).Return(errors.New("kafka unavailable"))
	mockRepo.On("UpdateStatus", mock.Anything, mock.AnythingOfType("string"), models.StatusFailed, mock.AnythingOfType("string")).Return(nil)

	response, err := service.ProcessNotification(req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "failed to queue notification")

	mockUserClient.AssertExpectations(t)
	mockTemplateClient.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockKafkaManager.AssertExpectations(t)
}

func TestOrchestrationService_UpdateNotificationStatus(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	ctx := context.Background()
	notificationID := "notif-123"
	status := models.StatusDelivered
	errorMsg := ""

	// Mock expectations
	mockRepo.On("UpdateStatus", ctx, notificationID, status, errorMsg).Return(nil)

	err := service.UpdateNotificationStatus(ctx, notificationID, status, errorMsg)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrchestrationService_UpdateNotificationStatus_WithError(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	ctx := context.Background()
	notificationID := "notif-123"
	status := models.StatusFailed
	errorMsg := "delivery failed"

	// Mock expectations
	mockRepo.On("UpdateStatus", ctx, notificationID, status, errorMsg).Return(nil)

	err := service.UpdateNotificationStatus(ctx, notificationID, status, errorMsg)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestOrchestrationService_ValidateChannelPreferences(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	tests := []struct {
		name             string
		notificationType models.NotificationType
		prefs            *models.UserPreferences
		expectedError    string
	}{
		{
			name:             "email enabled",
			notificationType: models.NotificationEmail,
			prefs:            &models.UserPreferences{Email: true, Push: false},
			expectedError:    "",
		},
		{
			name:             "email disabled",
			notificationType: models.NotificationEmail,
			prefs:            &models.UserPreferences{Email: false, Push: true},
			expectedError:    "email notifications disabled",
		},
		{
			name:             "push enabled",
			notificationType: models.NotificationPush,
			prefs:            &models.UserPreferences{Email: false, Push: true},
			expectedError:    "",
		},
		{
			name:             "push disabled",
			notificationType: models.NotificationPush,
			prefs:            &models.UserPreferences{Email: true, Push: false},
			expectedError:    "push notifications disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateChannelPreferences(tt.notificationType, tt.prefs)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestOrchestrationService_GetPriority(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	tests := []struct {
		name     string
		priority int
		expected string
	}{
		{"zero priority", 0, "normal"},
		{"low priority", 1, "low"},
		{"normal priority", 2, "normal"},
		{"high priority", 3, "high"},
		{"urgent priority", 4, "urgent"},
		{"unknown priority", 99, "normal"},
		{"negative priority", -1, "normal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.getPriority(tt.priority)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestOrchestrationService_CreateKafkaPayload_Email(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	notificationID := "notif-123"
	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationEmail,
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		Variables:        map[string]interface{}{"name": "John"},
		Priority:         3,
		Metadata:         map[string]interface{}{"source": "api"},
	}

	rendered := &models.RenderResponse{
		TemplateID: "welcome_email",
		Language:   "en",
		Version:    "1.0",
		Rendered: models.RenderedContent{
			Subject: "Welcome",
			Body: models.TemplateBody{
				HTML: "<h1>Welcome John</h1>",
				Text: "Welcome John",
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: []string{"name"},
	}

	payload := service.createKafkaPayload(notificationID, req, rendered)

	assert.NotNil(t, payload)
	assert.Equal(t, notificationID, payload.NotificationID)
	assert.Equal(t, "email", payload.NotificationType)
	assert.Equal(t, req.UserID, payload.UserID)
	assert.Equal(t, req.TemplateCode, payload.TemplateCode)
	assert.Equal(t, "high", payload.Priority)
	assert.Equal(t, "Welcome", payload.Subject)
	assert.Equal(t, "<h1>Welcome John</h1>", payload.Body)
	assert.Equal(t, "Welcome John", payload.TextBody)
	assert.Equal(t, req.Metadata, payload.Metadata)
}

func TestOrchestrationService_CreateKafkaPayload_Push(t *testing.T) {
	mockUserClient := new(MockUserClient)
	mockTemplateClient := new(MockTemplateClient)
	mockKafkaManager := new(MockKafkaManager)
	mockRepo := new(MockNotificationRepository)

	service := NewOrchestrationService(
		mockUserClient,
		mockTemplateClient,
		mockKafkaManager,
		mockRepo,
	)

	notificationID := "notif-123"
	req := &models.NotificationRequest{
		RequestID:        "req-123",
		NotificationType: models.NotificationPush,
		UserID:           "user-456",
		TemplateCode:     "push_notification",
		Variables:        map[string]interface{}{"message": "Hello"},
		Priority:         2,
	}

	rendered := &models.RenderResponse{
		TemplateID: "push_notification",
		Language:   "en",
		Version:    "1.0",
		Rendered: models.RenderedContent{
			Subject: "Push Notification",
			Body: models.TemplateBody{
				HTML: "",
				Text: "Hello",
			},
		},
		RenderedAt:    time.Now(),
		VariablesUsed: []string{"message"},
	}

	payload := service.createKafkaPayload(notificationID, req, rendered)

	assert.NotNil(t, payload)
	assert.Equal(t, notificationID, payload.NotificationID)
	assert.Equal(t, "push", payload.NotificationType)
	assert.Equal(t, req.UserID, payload.UserID)
	assert.Equal(t, req.TemplateCode, payload.TemplateCode)
	assert.Equal(t, "normal", payload.Priority)
	assert.Equal(t, "Push Notification", payload.Subject)
	assert.Equal(t, "Hello", payload.Body)
}
