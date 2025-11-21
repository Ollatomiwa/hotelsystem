package models

import (
	"time"

	"github.com/google/uuid"
)

type Payment struct {
	Id 	uuid.UUID `json:"id" db:"id"`
	BookingId uuid.UUID `json:"booking_id" db:"booking_id"`
	UserId uuid.UUID `json:"user_id" db:"user_id"`
	Amount float64 `json:"amount" db:"amount"`
	Currency string `json:"currency" db:"currency"`
	Status string `json:"status" db:"status"`
	PaymentMethod string `json:"payment_method" db:"paymeny_method"`
	PaymentProvider string `json:"payment_provider" db:"payment_provider"`
	ProviderTransaction_Id string `json:"provider_transaction_id" db:"provider_transaction_id"`
	Metadata string `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"creates_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CreatePaymentRequest struct {
	BookingId uuid.UUID `json:"booking_id" binding:"required"`
	UserId uuid.UUID `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
	Currency string `json:"currency" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	PaymentProvider string `json:"payment_provider" binding:"required"`
}

type PaymentResponse struct {
	Id 	uuid.UUID `json:"id" `
	BookingId uuid.UUID `json:"booking_id"`
	User_d uuid.UUID `json:"user_id"`
	Amount float64 `json:"amount"`
	Currency string `json:"currency"`
	Status string `json:"status"`
	PaymentMethod string `json:"payment_method"`
}