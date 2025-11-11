package mocks

import (
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
)

type UserServiceMock struct{}

func NewUserServiceMock() *UserServiceMock {
	return &UserServiceMock{}
}

func (m *UserServiceMock) GetPreferences(userID string) (*models.UserPreferences, error) {
	// Return mock user preferences
	return &models.UserPreferences{
		Email: true,
		Push:  true,
	}, nil
}
