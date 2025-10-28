//Core business operations
package services

import (
	"context"
	"fmt"
	"time"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories"
)

type NotificationService struct {
	repo repositories.NotificationRepository
}

//noewnotificationservice creates a new notification service
func NewNotificationService(repo repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		repo: repo,
	}
}

//implementing sendemail to handle the business logic for sending an email
func (s *NotificationService)SendEmail(ctx context.Context, req *models.SendEmailRequest) (*models.SendEmailResponse, error) {
	//creating notificaion record
	notification := &models.Notification {
		Id: generateId(),
		To: req.To,
		Subject: req.Subject,
		Body: req.Body,
		Status: models.StatusPending,
		Type:req.Type,
		RetryCount: 0,
		CreatedAt: time.Now(),
		SentAt: nil,
		Error: "",
	}

	//save to db
	err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification record: %w", err)
	}
	//for now we will just log that we would send an email, we will replace this with actual SMIP logic
	fmt.Printf("WOULD SEND EMAIL: TO: %s, Subject: %s\n", req.To, req.Subject)
	//simulaitng successgul sending for now
	err = s.repo.UpdateNotificationSent(ctx, notification.Id, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to update notification as sent:%w", err)
	}

	//return success message
	response := &models.SendEmailResponse{
		Id: notification.Id,
		Status: models.StatusSent,
		Message: "Email queued for delivery",
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	return response, nil
}

//generatedId creates unique ids for notifications
func generateId() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}

//getnotificationstatus will retrieves the status of a notification
func (s *NotificationService) GetNotificationStatus(ctx context.Context, id string)(*models.Notification, error){
	notification, err := s.repo.GetNotificationById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification status:%w", err)
	}
	return notification, nil
}