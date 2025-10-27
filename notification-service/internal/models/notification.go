package models

import (
	"time"
)

//Notification statuses
const (
	StatusPending = "pending"
	StatusSent = "sent"
	StatusFailed = "failed"
)

//Notification types
const (
	TypeBookingConfirmation = "booking_confirmation"
	TypeWelcomeEmail = "welcome_email"
	TypePaymentReceipt = "payment_receipt"
	TypePasswordReset = "password_reset"
)

//Notification represents an email notification in the system
type Notification struct {
	Id string `json:"id"`
	To string `json:"to"`
	Subject string `json:"subject"`
	Body string `json:"body"`
	Status string `json:"status"`
	Type string `json:"type"` //booking confirmation, welcome mail
	RetryCount int `json:"retry_count"`
	CreatedAt time.Time `json:"created_at"`
	SentAt *time.Time `json:"sent_at"`
	Error string `json:"error"`
}

