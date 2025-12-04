package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
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

func(r *PaymentRepository) GetPaymentById(ctx context.Context, paymentId uuid.UUID)( *models.Payment, error) {
	query := `
	SELECT id, booking_id, user_id, amount,currency, status, payment_method, payment_provider, 
	provider_transaction_id, metadata, created_At, updated_at WHERE id =$1`
	
	var payment models.Payment
	err := r.db.QueryRowContext(ctx, query,paymentId).Scan(&payment.Id, &payment.BookingId,
	&payment.UserId, &payment.Amount, &payment.Currency, &payment.Status, &payment.PaymentMethod,
	&payment.PaymentProvider, &payment.ProviderTransaction_Id, &payment.Metadata, &payment.CreatedAt, &payment.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("payment not found") 
	}

	return &payment, err
}

func(r *PaymentRepository) GetPaymentByBookingId(ctx context.Context, bookingId uuid.UUID)([]models.Payment, error){
	query := `
	SELECT id, booking_id, user_id, amount,currency, status, payment_method, payment_provider, 
	provider_transaction_id, metadata, created_At, updated_at WHERE booking_id =$
	`
	rows, err := r.db.QueryContext(ctx, query, bookingId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next(){
		var payment models.Payment
		err := rows.Scan(&payment.Id, &payment.BookingId,
	&payment.UserId, &payment.Amount, &payment.Currency, &payment.Status, &payment.PaymentMethod,
	&payment.PaymentProvider, &payment.ProviderTransaction_Id, &payment.Metadata, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err 
	}
	payments = append(payments, payment)
	}
	return payments, nil
}