package repository

import (
	"context"
	"database/sql"

	"github.com/ollatomwa/hotelsystem/payment-service/internals/models"
)

type PaymentRepository struct {
	db *sql.DB 
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository {db:db}
}

func(r *PaymentRepository) CreatePayment(ctx context.Context, payment *models.Payment) error {
	query := `INSERT INTO payments(id, booking_id, user_id, amount, currency, status, payment_method, payment_provider, provider_transaction_id, metadata, created_at, updated_at) 
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`

	_, err := r.db.ExecContext(ctx, query, 
		payment.Id,
		payment.BookingId,
		payment.UserId, 
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentProvider,
		payment.ProviderTransaction_Id,
		payment.Metadata,
		payment.CreatedAt,
		payment.UpdatedAt,
	)
	return err
}