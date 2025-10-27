// Database interfaces
package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
)

//notificationrepository defines the contract for notification data operations
type NotificationRepository interface {
	CreatNotification (ctx context.Context, notification *models.Notification) error
	GetNotificationById (ctx context.Context, id string) (*models.Notification, error)
	UpdateNotificationStatus (ctx context.Context, id string, status string) error
	UpdateNotificationSent (ctx context.Context, id string, sentAt time.Time) error
	UpdateNotificationFailed (ctx context.Context, id string, errorMsg string) error
}

//templaterepo defines the contract for email template operations
type TemplateRepository interface {
	GetTemplateByType (ctx context.Context, templateType string) (*models.Template, error)
	CreateTemplate (ctx context.Context, template *models.Template) error
}

//common repos errors
var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrTemplateNotFound = errors.New("template not found")
	ErrDuplicateId = errors.New("duplicate notifcation id")
)