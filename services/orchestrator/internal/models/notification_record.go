package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// NotificationRecord represents a notification record in the database
type NotificationRecord struct {
	ID               string             `json:"id" db:"id"`
	UserID           string             `json:"user_id" db:"user_id"`
	TemplateCode     string             `json:"template_code" db:"template_code"`
	NotificationType string             `json:"notification_type" db:"notification_type"`
	Status           NotificationStatus `json:"status" db:"status"`
	Priority         string             `json:"priority" db:"priority"`
	Variables        JSONB              `json:"variables" db:"variables"`
	Metadata         *JSONB             `json:"metadata,omitempty" db:"metadata"`
	ErrorMessage     *string            `json:"error_message,omitempty" db:"error_message"`
	CreatedAt        time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" db:"updated_at"`
	ScheduledFor     *time.Time         `json:"scheduled_for,omitempty" db:"scheduled_for"`
}

// JSONB is a custom type for PostgreSQL JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, j)
}
