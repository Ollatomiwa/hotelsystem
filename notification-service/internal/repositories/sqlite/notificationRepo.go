package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories"
)

//implement the interface
var _ repositories.NotificationRepository = (*NotificationRepo)(nil)

//notificationrepo will implements the notificationrepository for sqlite
type NotificationRepo struct {
	db *sql.DB
}

//newnotifrepo creates a new sqlite notif repo
func NewNotificationRepo(db *sql.DB) *NotificationRepo {
	return &NotificationRepo{
		db: db,
	}
}

//createnotif saves a new notification to the db
func (r *NotificationRepo) createNotification(ctx context.Context, notifcation *models.Notification) error {
	query := `
		INSERT INTO notifications
		(id, to_email, subject, body, status, type, retry_count, creates_at, sent_at, error)
		VALUES (?,?,?,?,?,?,?,?,?,?)
	`
	_, err := r.db.ExecContext(ctx, query, notifcation.Id,notifcation.To,
	notifcation.Subject,
	notifcation.Body,
	notifcation.Status,
	notifcation.Status,
	notifcation.Type,
	notifcation.RetryCount,
	notifcation.CreatedAt,
	notifcation.SentAt,
	notifcation.Error,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

//methods
//getnotification by id
func (r *NotificationRepo) GetNotificationById(id string) (*models.Notification, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		SELECT id, to_email, subject, body, status, types, retry_count, created_at, sent_at, error 
		FROM notifications
		WHERE id = ?
	`
	var notification models.Notification
	var sentAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query,id).Scan(
		&notification.Id,
		&notification.To,
		notification.Subject,
		&notification.Body,
		&notification.Status,
		&notification.Type,
		&notification.RetryCount,
		&notification.CreatedAt,
		&sentAt,
		&notification.Error,
	)
	if err == sql.ErrNoRows{
		return nil, repositories.ErrNotificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	//convert sql.nulltime to *time.time
	if sentAt.Valid {
		notification.SentAt = &sentAt.Time
	}
	return &notification, nil
}
