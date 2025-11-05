package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

// SendEmailRequest matches what your Notification Service expects
type SendEmailRequest struct {
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
	Type    string `json:"type,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// SendBookingConfirmation sends a notification when a booking is confirmed
func (c *Client) SendBookingConfirmation(ctx context.Context, userID string, bookingData map[string]interface{}) error {
	// For now, use userID as email. In production, you'd look up the user's email
	email := userID + "@example.com" // Temporary - replace with actual email lookup
	
	subject := "Booking Confirmed!"
	body := fmt.Sprintf(
		"Your booking has been confirmed!\n\n"+
		"Room: %s\n"+
		"Check-in: %s\n"+
		"Check-out: %s\n"+
		"Guests: %d\n"+
		"Total Amount: $%.2f\n\n"+
		"Thank you for choosing our hotel!",
		bookingData["room_number"],
		bookingData["check_in"],
		bookingData["check_out"],
		bookingData["guests"],
		bookingData["total_amount"],
	)

	req := SendEmailRequest{
		To:      email,
		Subject: subject,
		Body:    body,
		Type:    "booking_confirmation",
		Data:    bookingData,
	}

	return c.sendEmailNotification(ctx, req)
}

// SendBookingCancellation sends a notification when a booking is cancelled
func (c *Client) SendBookingCancellation(ctx context.Context, userID string, bookingData map[string]interface{}) error {
	// For now, use userID as email. In production, you'd look up the user's email
	email := userID + "@example.com" // Temporary - replace with actual email lookup
	
	subject := " Booking Cancelled"
	body := fmt.Sprintf(
		"Your booking has been cancelled.\n\n"+
		"Booking ID: %s\n"+
		"Room: %s\n"+
		"Original Check-in: %s\n\n"+
		"If this was a mistake, please contact us.",
		bookingData["booking_id"],
		bookingData["room_number"],
		bookingData["check_in"],
	)

	req := SendEmailRequest{
		To:      email,
		Subject: subject,
		Body:    body,
		Type:    "booking_cancellation",
		Data:    bookingData,
	}

	return c.sendEmailNotification(ctx, req)
}

func (c *Client) sendEmailNotification(ctx context.Context, req SendEmailRequest) error {
	url := c.baseURL + "/api/v1/notifications/email"

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal email request: %w", err)
	}

	fmt.Printf(" DEBUG: Sending email to: %s\n", url)
	fmt.Printf(" DEBUG: Email request: %s\n", string(body))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Booking-Service/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		fmt.Printf(" DEBUG: HTTP error: %v\n", err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)
	fmt.Printf("DEBUG: Response status: %d\n", resp.StatusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(responseBody))

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("notification service returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	fmt.Printf("Email sent successfully to: %s\n", req.To)
	return nil
}

// HealthCheck verifies the notification service is reachable
func (c *Client) HealthCheck(ctx context.Context) error {
	url := c.baseURL + "/api/v1/notifications/email"
	
	// Send a minimal test email request
	testReq := SendEmailRequest{
		To:      "health@check.com",
		Subject: "Health Check",
		Body:    "Service health check",
	}

	body, err := json.Marshal(testReq)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("notification service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("notification service returned server error %d", resp.StatusCode)
	}

	return nil
}