package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"

    "github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
    "github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories"
)

type NotificationRepo struct {
    db *sql.DB
}

func NewNotificationRepo(db *sql.DB) *NotificationRepo {
    return &NotificationRepo{db: db}
}

func (r *NotificationRepo) CreateNotification(ctx context.Context, notification *models.Notification) error {
    query := `
        INSERT INTO notifications 
        (id, to_email, subject, body, status, type, retry_count, sent_at, error) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    
    _, err := r.db.ExecContext(ctx, query,
        notification.Id,
        notification.To,
        notification.Subject,
        notification.Body,
        notification.Status,
        notification.Type,
        notification.RetryCount,
        notification.SentAt,  // Now just sent_at, no created_at
        notification.Error,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create notification: %w", err)
    }
    
    return nil
}

func (r *NotificationRepo) GetNotificationById(ctx context.Context, id string) (*models.Notification, error) {
    query := `
        SELECT id, to_email, subject, body, status, type, retry_count, sent_at, error
        FROM notifications 
        WHERE id = $1
    `
    
    var notification models.Notification
    var sentAtStr sql.NullString
    var errorStr sql.NullString
    
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &notification.Id,
        &notification.To,
        &notification.Subject,
        &notification.Body,
        &notification.Status,
        &notification.Type,
        &notification.RetryCount,
        &sentAtStr,
        &errorStr,
    )
    
    if err == sql.ErrNoRows {
        return nil, repositories.ErrNotificationNotFound
    }
    
    if err != nil {
        return nil, fmt.Errorf("failed to get notification: %w", err)
    }
    
    // Convert string to *time.Time if sent_at exists
    if sentAtStr.Valid && sentAtStr.String != "" {
        sentTime, err := time.Parse(time.RFC3339, sentAtStr.String)
        if err != nil {
            notification.SentAt = nil
        } else {
            notification.SentAt = &sentTime
        }
    } else {
        notification.SentAt = nil
    }
    
    // Handle NULL error field
    if errorStr.Valid {
        notification.Error = errorStr.String
    } else {
        notification.Error = ""
    }
    
    return &notification, nil
}

func (r *NotificationRepo) UpdateNotificationStatus(ctx context.Context, id string, status string) error {
    query := `UPDATE notifications SET status = $1 WHERE id = $2`
    
    result, err := r.db.ExecContext(ctx, query, status, id)
    if err != nil {
        return fmt.Errorf("failed to update notification status: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return repositories.ErrNotificationNotFound
    }
    
    return nil
}

func (r *NotificationRepo) UpdateNotificationSent(ctx context.Context, id string, sentAt time.Time) error {
    query := `
        UPDATE notifications 
        SET status = $1, sent_at = $2, retry_count = retry_count + 1 
        WHERE id = $3
    `
    
    // Convert time to string for PostgreSQL
    sentAtStr := sentAt.Format(time.RFC3339)
    
    result, err := r.db.ExecContext(ctx, query, models.StatusSent, sentAtStr, id)
    if err != nil {
        return fmt.Errorf("failed to update notification as sent: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return repositories.ErrNotificationNotFound
    }
    
    return nil
}

func (r *NotificationRepo) UpdateNotificationFailed(ctx context.Context, id string, errorMsg string) error {
    query := `
        UPDATE notifications 
        SET status = $1, error = $2, retry_count = retry_count + 1 
        WHERE id = $3
    `
    
    result, err := r.db.ExecContext(ctx, query, models.StatusFailed, errorMsg, id)
    if err != nil {
        return fmt.Errorf("failed to update notification as failed: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rowsAffected == 0 {
        return repositories.ErrNotificationNotFound
    }
    
    return nil
}