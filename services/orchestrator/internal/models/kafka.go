package models

import "time"

// KafkaNotificationPayload is the message structure sent to Kafka
// This is what the Email and Push services will consume
type KafkaNotificationPayload struct {
	NotificationID   string                 `json:"notification_id"`
	NotificationType string                 `json:"notification_type"`
	UserID           string                 `json:"user_id"`
	TemplateCode     string                 `json:"template_code"`
	Subject          string                 `json:"subject,omitempty"`
	Body             string                 `json:"body"`
	TextBody         string                 `json:"text_body,omitempty"`
	Priority         string                 `json:"priority"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`

	RetryCount  int       `json:"retry_count,omitempty"`
	LastRetryAt time.Time `json:"last_retry_at,omitempty"`
}
