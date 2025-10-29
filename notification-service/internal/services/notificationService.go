//Core business operations
package services

import (
	"context"
	"fmt"
	"time"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/models"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/email"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/ratelimiter"
)

type NotificationService struct {
	repo repositories.NotificationRepository
	emailSender *email.SMTPSender
	rateLimiter *ratelimiter.RateLimiter

}

//newnotificationservice creates a new notification service
func NewNotificationService(
	repo repositories.NotificationRepository,
	emailSender *email.SMTPSender,
	rateLimiter *ratelimiter.RateLimiter,
	) *NotificationService {
	return &NotificationService{
		repo: repo,
		emailSender: emailSender,
		rateLimiter: rateLimiter,
	}
}

//implementing sendemail to handle the business logic for sending an email
func (s *NotificationService)SendEmail(ctx context.Context, req *models.SendEmailRequest) (*models.SendEmailResponse, error) {
	
	//check rate limits for the recipent
	if !s.rateLimiter.Allow(req.To) {
		return nil, fmt.Errorf("rate limit exceeded for email: %s", req.To)
	}
	
	//creating notificaion record
	notification := &models.Notification {
		Id: generateId(),
		To: req.To,
		Subject: req.Subject,
		Body: req.Body,
		Status: models.StatusPending,
		Type:req.Type,
		RetryCount: 0,
		SentAt: nil,
		Error: "",
	}

	//save to db
	err := s.repo.CreateNotification(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to create notification record: %w", err)
	}

	//send the actual mail
	err = s.emailSender.SendEmail(req.To, req.Subject, req.Body)
	if err != nil {
		//update notif as failed
		updateErr := s.repo.UpdateNotificationFailed(ctx, notification.Id, err.Error())

		if updateErr != nil {
			return nil, fmt.Errorf("failed to send email: %w, and failed to update status: %v", err, updateErr)
		}
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	//update notification as sent
	now := time.Now()
	err = s.repo.UpdateNotificationSent(ctx, notification.Id, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update notification as sent: %w", err)
	}



	// //for now we will just log that we would send an email, we will replace this with actual SMIP logic
	// fmt.Printf("WOULD SEND EMAIL: TO: %s, Subject: %s\n", req.To, req.Subject)
	// //simulaitng successgul sending for now
	// err = s.repo.UpdateNotificationSent(ctx, notification.Id, time.Now())
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to update notification as sent:%w", err)
	// }


	//return success message
	response := &models.SendEmailResponse{
		Id: notification.Id,
		Status: models.StatusSent,
		Message: "Email sent successfully",
		TimeStamp: time.Now().Format(time.RFC3339),
	}

	return response, nil
}


//getnotificationstatus will retrieves the status of a notification
func (s *NotificationService) GetNotificationStatus(ctx context.Context, id string)(*models.Notification, error){
	notification, err := s.repo.GetNotificationById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification status:%w", err)
	}
	return notification, nil
}

//getratelimitstatus to return remaining reqs for an email
func (s *NotificationService) GetRateLimitStatus(email string) int {
	return s.rateLimiter.GetRemainingRequests(email)
}

//generatedId creates unique ids for notifications
func generateId() string {
	return fmt.Sprintf("notif_%d", time.Now().UnixNano())
}