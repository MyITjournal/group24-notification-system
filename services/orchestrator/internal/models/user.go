package models

import "time"

// UserData contains a user's data
type UserData struct {
	Name string        `json:"name"`
	Link string        `json:"link"`
	Meta *UserMetaData `json:"meta"`
}

// UserMetaData represents metadata for a single user

// TODO: find out about user meta data
type UserMetaData struct {
}

// UserDevice represents a registered device for push notifications.
type UserDevice struct {
	Token     string    `json:"token" binding:"required"`
	Platform  string    `json:"platform" binding:"required,oneof=ios android web"`
	CreatedAt time.Time `json:"created_at"`
}

// UserCreationRequest is the payload for creating a new user.
type UserCreationRequest struct {
	Name        string          `json:"name" binding:"required"`
	Email       string          `json:"email" binding:"required,email"`
	PushToken   string          `json:"push_token,omitempty"`
	Preferences UserPreferences `json:"preferences"`
	Password    string          `json:"password" binding:"required,min=8"`
}

// UserPreferences holds all notification preferences for a user.
type UserPreferences struct {
	Email bool `json:"email_enabled"`
	Push  bool `json:"push_enabled"`
}
