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
func (r *NotificationRepo) CreateNotification(ctx context.Context, notification *models.Notification) error {
	query := `
		INSERT INTO notifications
		(id, to_email, subject, body, status, type, retry_count, sent_at, error)
		VALUES (?,?,?,?,?,?,?,?,?)
	`
	var sentAtStr interface{}
	if notification.SentAt != nil {
		sentAtStr = notification.SentAt.Format(time.RFC3339)
	} else {
		sentAtStr = nil
	}

	//ensuring err is nver null
	errorValue := notification.Error
	if errorValue == "" {
		errorValue = ""
	}

	_, err := r.db.ExecContext(ctx, query, notification.Id,notification.To,
	notification.Subject,
	notification.Body,
	notification.Status,
	notification.Status,
	notification.Type,
	notification.RetryCount,
	sentAtStr,
	errorValue,
	)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}
	return nil
}

//methods
//getnotification by id
func (r *NotificationRepo) GetNotificationById(ctx context.Context, id string) (*models.Notification, error) {

	query := `
		SELECT id, to_email, subject, body, status, type, retry_count, sent_at, error 
		FROM notifications
		WHERE id = ?
	`
	var notification models.Notification
	var sentAt sql.NullTime
	var sentAtStr sql.NullString
	var errorStr sql.NullString

	err := r.db.QueryRowContext(ctx, query,id).Scan(
		&notification.Id,
		&notification.To,
		&notification.Subject,
		&notification.Body,
		&notification.Status,
		&notification.Type,
		&notification.RetryCount,
		&sentAtStr, //scan as string first
		&errorStr, //scan as NUllstring instead of direct string
	)
	if err == sql.ErrNoRows{
		return nil, repositories.ErrNotificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	//convert sql.nulltime to *time.time if sentat exists
	if sentAt.Valid && sentAtStr.String != "" {
		sentTime, err := time.Parse(time.RFC3339, sentAtStr.String)
		if err != nil {
			notification.SentAt = nil
		} else {
			notification.SentAt = &sentTime
		}
	} else {
		notification.SentAt = nil
	}

	//handling null error filed
	if errorStr.Valid {
		notification.Error = errorStr.String
	} else {
		notification.Error = ""
	}
	return &notification, nil

}

//update notification status
func (r *NotificationRepo) UpdateNotificationStatus(ctx context.Context, id string, status string) error {

	query := `UPDATE notification SET status = ? WHERE id = ?`
	results, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affect: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}

//update notification sent to updates the notifs with sent timestamp and status
func (r *NotificationRepo) UpdateNotificationSent(ctx context.Context, id string, sentAt time.Time) error {
	
	query := `UPDATE notifications SET Status = ?, sent_at =?, retry_count = retry_count + 1 WHERE id= ?`

	sentAtStr := sentAt.Format(time.RFC3339)

	results, err := r.db.ExecContext(ctx, query, models.StatusSent, sentAtStr, id)
	if err != nil {
		return fmt.Errorf("failed to update notification as sent: %w", err)
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affect: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}

//update notification failed updates notifs with error info
func ( r *NotificationRepo) UpdateNotificationFailed(ctx context.Context, id string, errorMsg string) error {

	query := `UPDATE notifications SET status = ?, error = ?, retry_count = retry_count + 1 WHERE id = ?`

	results, err := r.db.ExecContext(ctx, query, models.StatusFailed, errorMsg, id)
	if err != nil {
		return fmt.Errorf("failed to update notification as failed: %w", err)
	}

	rowsAffected, err := results.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affect: %w", err)
	}

	if rowsAffected == 0 {
		return repositories.ErrNotificationNotFound
	}
	return nil
}