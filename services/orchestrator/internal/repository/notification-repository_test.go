package repository

// Note: This test file requires the sqlmock dependency.
// Run: go get github.com/DATA-DOG/go-sqlmock

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/models"
	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain initializes the logger before running tests
func TestMain(m *testing.M) {
	if err := logger.Initialize("info", "console"); err != nil {
		panic("Failed to initialize logger for tests: " + err.Error())
	}
	defer logger.Sync()
	code := m.Run()
	os.Exit(code)
}

func setupMockDB(t *testing.T) (*database.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	return &database.DB{DB: db}, mock
}

func TestNewNotificationRepository(t *testing.T) {
	db, _ := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)
	assert.NotNil(t, repo)
}

func TestNotificationRepository_Create_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notification := &models.NotificationRecord{
		ID:               "notif-123",
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "John"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs(
			notification.ID,
			notification.UserID,
			notification.TemplateCode,
			notification.NotificationType,
			notification.Status,
			notification.Priority,
			sqlmock.AnyArg(), // variablesJSON
			sqlmock.AnyArg(), // metadataJSON (nil)
			sqlmock.AnyArg(), // errorMessage (nil)
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // scheduled_for (nil)
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.Create(ctx, notification)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_Create_WithMetadata(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	metadata := models.JSONB{"source": "api", "campaign": "welcome"}
	notification := &models.NotificationRecord{
		ID:               "notif-123",
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "John"},
		Metadata:         &metadata,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs(
			notification.ID,
			notification.UserID,
			notification.TemplateCode,
			notification.NotificationType,
			notification.Status,
			notification.Priority,
			sqlmock.AnyArg(), // variablesJSON
			sqlmock.AnyArg(), // metadataJSON
			sqlmock.AnyArg(), // errorMessage
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // scheduled_for
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.Create(ctx, notification)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_Create_WithScheduledFor(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	scheduledFor := time.Now().Add(24 * time.Hour)
	notification := &models.NotificationRecord{
		ID:               "notif-123",
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "John"},
		ScheduledFor:     &scheduledFor,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs(
			notification.ID,
			notification.UserID,
			notification.TemplateCode,
			notification.NotificationType,
			notification.Status,
			notification.Priority,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	ctx := context.Background()
	err := repo.Create(ctx, notification)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_Create_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notification := &models.NotificationRecord{
		ID:               "notif-123",
		UserID:           "user-456",
		TemplateCode:     "welcome_email",
		NotificationType: "email",
		Status:           models.StatusPending,
		Priority:         "normal",
		Variables:        models.JSONB{"name": "John"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	mock.ExpectExec(`INSERT INTO notifications`).
		WithArgs(
			notification.ID,
			notification.UserID,
			notification.TemplateCode,
			notification.NotificationType,
			notification.Status,
			notification.Priority,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err := repo.Create(ctx, notification)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create notification record")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	variablesJSON, _ := json.Marshal(models.JSONB{"name": "John"})

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		notificationID,
		"user-456",
		"welcome_email",
		"email",
		models.StatusPending,
		"normal",
		variablesJSON,
		nil,                          // metadata
		sql.NullString{Valid: false}, // error_message
		time.Now(),
		time.Now(),
		nil, // scheduled_for
	)

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnRows(rows)

	ctx := context.Background()
	record, err := repo.GetByID(ctx, notificationID)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.Equal(t, notificationID, record.ID)
	assert.Equal(t, "user-456", record.UserID)
	assert.Equal(t, "welcome_email", record.TemplateCode)
	assert.Equal(t, models.StatusPending, record.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByID_WithMetadata(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	variablesJSON, _ := json.Marshal(models.JSONB{"name": "John"})
	metadataJSON, _ := json.Marshal(models.JSONB{"source": "api"})

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		notificationID,
		"user-456",
		"welcome_email",
		"email",
		models.StatusPending,
		"normal",
		variablesJSON,
		metadataJSON,
		sql.NullString{Valid: false},
		time.Now(),
		time.Now(),
		nil,
	)

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnRows(rows)

	ctx := context.Background()
	record, err := repo.GetByID(ctx, notificationID)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.NotNil(t, record.Metadata)
	assert.Equal(t, "api", (*record.Metadata)["source"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByID_WithErrorMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	variablesJSON, _ := json.Marshal(models.JSONB{"name": "John"})
	errorMsg := "delivery failed"

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		notificationID,
		"user-456",
		"welcome_email",
		"email",
		models.StatusFailed,
		"normal",
		variablesJSON,
		nil,
		sql.NullString{Valid: true, String: errorMsg},
		time.Now(),
		time.Now(),
		nil,
	)

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnRows(rows)

	ctx := context.Background()
	record, err := repo.GetByID(ctx, notificationID)

	assert.NoError(t, err)
	assert.NotNil(t, record)
	assert.NotNil(t, record.ErrorMessage)
	assert.Equal(t, errorMsg, *record.ErrorMessage)
	assert.Equal(t, models.StatusFailed, record.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByID_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnError(sql.ErrNoRows)

	ctx := context.Background()
	record, err := repo.GetByID(ctx, notificationID)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "notification not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByID_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"

	mock.ExpectQuery(`SELECT (.+) FROM notifications WHERE id = \$1`).
		WithArgs(notificationID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	record, err := repo.GetByID(ctx, notificationID)

	assert.Error(t, err)
	assert.Nil(t, record)
	assert.Contains(t, err.Error(), "failed to get notification")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_UpdateStatus_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	status := models.StatusDelivered

	mock.ExpectExec(`UPDATE notifications`).
		WithArgs(status, sqlmock.AnyArg(), notificationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err := repo.UpdateStatus(ctx, notificationID, status, "")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_UpdateStatus_WithErrorMessage(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	status := models.StatusFailed
	errorMsg := "delivery failed"

	mock.ExpectExec(`UPDATE notifications`).
		WithArgs(status, sqlmock.AnyArg(), notificationID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	ctx := context.Background()
	err := repo.UpdateStatus(ctx, notificationID, status, errorMsg)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_UpdateStatus_NotFound(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	status := models.StatusDelivered

	mock.ExpectExec(`UPDATE notifications`).
		WithArgs(status, sqlmock.AnyArg(), notificationID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

	ctx := context.Background()
	err := repo.UpdateStatus(ctx, notificationID, status, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_UpdateStatus_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	notificationID := "notif-123"
	status := models.StatusDelivered

	mock.ExpectExec(`UPDATE notifications`).
		WithArgs(status, sqlmock.AnyArg(), notificationID).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	err := repo.UpdateStatus(ctx, notificationID, status, "")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update notification status")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByUserID_Success(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	userID := "user-456"
	limit := 10
	offset := 0

	variablesJSON1, _ := json.Marshal(models.JSONB{"name": "John"})
	variablesJSON2, _ := json.Marshal(models.JSONB{"name": "Jane"})

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		"notif-1",
		userID,
		"welcome_email",
		"email",
		models.StatusPending,
		"normal",
		variablesJSON1,
		nil,
		sql.NullString{Valid: false},
		time.Now(),
		time.Now(),
		nil,
	).AddRow(
		"notif-2",
		userID,
		"push_notification",
		"push",
		models.StatusDelivered,
		"high",
		variablesJSON2,
		nil,
		sql.NullString{Valid: false},
		time.Now(),
		time.Now(),
		nil,
	)

	mock.ExpectQuery(`SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	ctx := context.Background()
	records, err := repo.GetByUserID(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.Len(t, records, 2)
	assert.Equal(t, "notif-1", records[0].ID)
	assert.Equal(t, "notif-2", records[1].ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByUserID_EmptyResult(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	userID := "user-456"
	limit := 10
	offset := 0

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	})

	mock.ExpectQuery(`SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	ctx := context.Background()
	records, err := repo.GetByUserID(ctx, userID, limit, offset)

	assert.NoError(t, err)
	// When no rows are found, Go returns nil slice, which is valid
	// We just check the length is 0 (nil slice has length 0)
	assert.Len(t, records, 0)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByUserID_WithPagination(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	userID := "user-456"
	limit := 5
	offset := 10

	variablesJSON, _ := json.Marshal(models.JSONB{"name": "John"})

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		"notif-1",
		userID,
		"welcome_email",
		"email",
		models.StatusPending,
		"normal",
		variablesJSON,
		nil,
		sql.NullString{Valid: false},
		time.Now(),
		time.Now(),
		nil,
	)

	mock.ExpectQuery(`SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	ctx := context.Background()
	records, err := repo.GetByUserID(ctx, userID, limit, offset)

	assert.NoError(t, err)
	assert.NotNil(t, records)
	assert.Len(t, records, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByUserID_DatabaseError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	userID := "user-456"
	limit := 10
	offset := 0

	mock.ExpectQuery(`SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnError(sql.ErrConnDone)

	ctx := context.Background()
	records, err := repo.GetByUserID(ctx, userID, limit, offset)

	assert.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "failed to query notifications")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNotificationRepository_GetByUserID_ScanError(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewNotificationRepository(db)

	userID := "user-456"
	limit := 10
	offset := 0

	// Create rows with invalid data to cause scan error
	rows := sqlmock.NewRows([]string{
		"id", "user_id", "template_code", "notification_type", "status", "priority",
		"variables", "metadata", "error_message", "created_at", "updated_at", "scheduled_for",
	}).AddRow(
		"notif-1",
		"user-456",
		"welcome_email",
		"email",
		models.StatusPending,
		"normal",
		"invalid-json", // This will cause unmarshal error
		nil,
		sql.NullString{Valid: false},
		time.Now(),
		time.Now(),
		nil,
	)

	mock.ExpectQuery(`SELECT id, user_id, template_code, notification_type, status, priority,
		       variables, metadata, error_message, created_at, updated_at, scheduled_for
		FROM notifications
		WHERE user_id = \$1
		ORDER BY created_at DESC
		LIMIT \$2 OFFSET \$3`).
		WithArgs(userID, limit, offset).
		WillReturnRows(rows)

	ctx := context.Background()
	records, err := repo.GetByUserID(ctx, userID, limit, offset)

	assert.Error(t, err)
	assert.Nil(t, records)
	assert.Contains(t, err.Error(), "failed to unmarshal variables")
	assert.NoError(t, mock.ExpectationsWereMet())
}
