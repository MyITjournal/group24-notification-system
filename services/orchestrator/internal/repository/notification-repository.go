package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"go.uber.org/zap"
)

// NotificationRepository defines the interface for notification persistence operations
type NotificationRepository interface {
	Create(ctx context.Context, notification *models.NotificationRecord) error
	GetByID(ctx context.Context, id string) (*models.NotificationRecord, error)
	UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.NotificationRecord, error)
}

// postgresNotificationRepository implements NotificationRepository using PostgreSQL
type postgresNotificationRepository struct {
	db *database.DB
}

// NewNotificationRepository creates a new PostgreSQL notification repository
func NewNotificationRepository(db *database.DB) NotificationRepository {
	return &postgresNotificationRepository{db: db}
}

// Create inserts a new notification record into the database
func (r *postgresNotificationRepository) Create(ctx context.Context, notification *models.NotificationRecord) error {
	query := `
		INSERT INTO notifications (
			id, user_id, template_code, notification_type, status, priority,
			variables, metadata, error_message, created_at, updated_at, scheduled_for
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	variablesJSON, err := json.Marshal(notification.Variables)
	if err != nil {
		return fmt.Errorf("failed to marshal variables: %w", err)
	}

	var metadataJSON []byte
	if notification.Metadata != nil {
		metadataJSON, err = json.Marshal(notification.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	now := time.Now()
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = now
	}
	if notification.UpdatedAt.IsZero() {
		notification.UpdatedAt = now
	}

	_, err = r.db.ExecContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.TemplateCode,
		notification.NotificationType,
		notification.Status,
		notification.Priority,
		variablesJSON,
		metadataJSON,
		notification.ErrorMessage,
		notification.CreatedAt,
		notification.UpdatedAt,
		notification.ScheduledFor,
	)

	if err != nil {
		logger.Log.Error("Failed to create notification record",
			zap.String("notification_id", notification.ID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	logger.Log.Debug("Notification record created",
		zap.String("notification_id", notification.ID),
		zap.String("user_id", notification.UserID),
	)

	return nil
}

// GetByID retrieves a notification record by its ID
func (r *postgresNotificationRepository) GetByID(ctx context.Context, id string) (*models.NotificationRecord, error) {
	query := `
		SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE id = $1
	`

	var record models.NotificationRecord
	var variablesJSON, metadataJSON []byte
	var errorMsg sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.UserID,
		&record.TemplateCode,
		&record.NotificationType,
		&record.Status,
		&record.Priority,
		&variablesJSON,
		&metadataJSON,
		&errorMsg,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.ScheduledFor,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("notification not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(variablesJSON, &record.Variables); err != nil {
		return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
	}

	if len(metadataJSON) > 0 {
		var metadata models.JSONB
		if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
		record.Metadata = &metadata
	}

	if errorMsg.Valid {
		record.ErrorMessage = &errorMsg.String
	}

	return &record, nil
}

// UpdateStatus updates the status of a notification record
func (r *postgresNotificationRepository) UpdateStatus(ctx context.Context, id string, status models.NotificationStatus, errorMsg string) error {
	query := `
		UPDATE notifications
		SET status = $1, error_message = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	var errorMsgPtr *string
	if errorMsg != "" {
		errorMsgPtr = &errorMsg
	}

	result, err := r.db.ExecContext(ctx, query, status, errorMsgPtr, id)
	if err != nil {
		logger.Log.Error("Failed to update notification status",
			zap.String("notification_id", id),
			zap.String("status", string(status)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found: %s", id)
	}

	logger.Log.Debug("Notification status updated",
		zap.String("notification_id", id),
		zap.String("status", string(status)),
	)

	return nil
}

// GetByUserID retrieves notifications for a specific user with pagination
func (r *postgresNotificationRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*models.NotificationRecord, error) {
	query := `
		SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	var records []*models.NotificationRecord
	for rows.Next() {
		var record models.NotificationRecord
		var variablesJSON, metadataJSON []byte
		var errorMsg sql.NullString

		err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.TemplateCode,
			&record.NotificationType,
			&record.Status,
			&record.Priority,
			&variablesJSON,
			&metadataJSON,
			&errorMsg,
			&record.CreatedAt,
			&record.UpdatedAt,
			&record.ScheduledFor,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(variablesJSON, &record.Variables); err != nil {
			return nil, fmt.Errorf("failed to unmarshal variables: %w", err)
		}

		if len(metadataJSON) > 0 {
			var metadata models.JSONB
			if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
			record.Metadata = &metadata
		}

		if errorMsg.Valid {
			record.ErrorMessage = &errorMsg.String
		}

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating notifications: %w", err)
	}

	return records, nil
}
