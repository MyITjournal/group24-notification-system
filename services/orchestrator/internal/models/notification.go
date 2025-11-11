package models

import "time"

type NotificationRequest struct {
	ID               string                 `json:"id" binding:"required"`
	NotificationType NotificationType       `json:"notification_type" binding:"required,oneof=email push"`
	UserID           string                 `json:"user_id" binding:"required"`
	TemplateCode     string                 `json:"template_code" binding:"required"`
	Variables        map[string]interface{} `json:"variables"`
	Priority         string                 `json:"priority,omitempty"`
	ScheduledFor     *time.Time             `json:"scheduled_for,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type NotificationType string

const (
	NotificationEmail NotificationType = "email"
	NotificationPush  NotificationType = "push"
)

type NotificationResponse struct {
	NotificationID string             `json:"notification_id"`
	Status         NotificationStatus `json:"status"`
	Timestamp      time.Time          `json:"timestamp"`
	Error          string             `json:"error,omitempty"`
}

type NotificationStatus string

const (
	StatusDelivered NotificationStatus = "delivered"
	StatusPending   NotificationStatus = "pending"
	StatusFailed    NotificationStatus = "failed"
)

type Notification struct {
	NotificationType string                 `json:"notification_type"`
	UserID           string                 `json:"user_id"`
	TemplateCode     string                 `json:"template_code"`
	Variables        map[string]interface{} `json:"variables"`
	RequestID        string                 `json:"request_id"`
	Priority         string                 `json:"priority"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}
