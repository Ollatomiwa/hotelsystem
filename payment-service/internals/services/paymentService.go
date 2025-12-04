package service

import (
	"encoding/json"
	"fmt"
	"payment-service/internal/models"
	"payment-service/internal/repository"
	"payment-service/pkg/paystack"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentService struct {
	repo           *repository.Repository
	paystackClient *paystack.Client
	logger         *logrus.Logger
}

func NewPaymentService(repo *repository.Repository, paystackClient *paystack.Client, logger *logrus.Logger) *PaymentService {
	return &PaymentService{
		repo:           repo,
		paystackClient: paystackClient,
		logger:         logger,
	}
}

// InitializePayment initializes a new payment
func (s *PaymentService) InitializePayment(req *models.InitializePaymentRequest) (*models.Transaction, error) {
	// Check for idempotency
	if req.IdempotencyKey != "" {
		existingTx, err := s.repo.GetTransactionByIdempotencyKey(req.IdempotencyKey)
		if err == nil {
			s.logger.Infof("Returning existing transaction for idempotency key: %s", req.IdempotencyKey)
			return existingTx, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
	}

	// Generate reference if not provided
	if req.Reference == "" {
		req.Reference = "TXN_" + uuid.New().String()
	}

	// Set currency default
	if req.Currency == "" {
		req.Currency = "NGN"
	}

	// Get or create customer
	customer, err := s.repo.GetOrCreateCustomer(req.Email, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get/create customer: %w", err)
	}

	// Initialize with Paystack
	paystackResp, err := s.paystackClient.InitializeTransaction(req)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment with Paystack: %w", err)
	}

	// Serialize metadata
	metadataJSON, _ := json.Marshal(req.Metadata)

	// Create transaction record
	transaction := &models.Transaction{
		Reference:      paystackResp.Data.Reference,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         models.StatusPending,
		CustomerEmail:  req.Email,
		CustomerName:   customer.Name,
		Metadata:       string(metadataJSON),
		AuthURL:        paystackResp.Data.AuthorizationURL,
		AccessCode:     paystackResp.Data.AccessCode,
		IdempotencyKey: req.IdempotencyKey,
	}

	if err := s.repo.CreateTransaction(transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	s.logger.Infof("Payment initialized: reference=%s, amount=%d", transaction.Reference, transaction.Amount)
	return transaction, nil
}

// VerifyPayment verifies a payment with Paystack
func (s *PaymentService) VerifyPayment(reference string) (*models.Transaction, error) {
	// Get transaction from database
	transaction, err := s.repo.GetTransactionByReference(reference)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// If already successful, return it
	if transaction.Status == models.StatusSuccess {
		return transaction, nil
	}

	// Verify with Paystack
	paystackResp, err := s.paystackClient.VerifyTransaction(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to verify with Paystack: %w", err)
	}

	// Update transaction status
	transaction.Status = s.mapPaystackStatus(paystackResp.Data.Status)
	transaction.PaystackRef = paystackResp.Data.Reference

	if err := s.repo.UpdateTransaction(transaction); err != nil {
		return nil, fmt.Errorf("failed to update transaction: %w", err)
	}

	// Update customer info if available
	if paystackResp.Data.Customer.CustomerCode != "" {
		customer, err := s.repo.GetCustomerByEmail(transaction.CustomerEmail)
		if err == nil {
			customer.CustomerCode = paystackResp.Data.Customer.CustomerCode
			_ = s.repo.UpdateCustomer(customer)
		}
	}

	s.logger.Infof("Payment verified: reference=%s, status=%s", reference, transaction.Status)
	return transaction, nil
}

// GetTransaction retrieves a transaction by ID
func (s *PaymentService) GetTransaction(id uint) (*models.Transaction, error) {
	return s.repo.GetTransactionByID(id)
}

// ListTransactions lists transactions with pagination
func (s *PaymentService) ListTransactions(page, pageSize int, status, email string) ([]models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return s.repo.ListTransactions(page, pageSize, status, email)
}

// GetCustomer retrieves a customer by email
func (s *PaymentService) GetCustomer(email string) (*models.Customer, error) {
	return s.repo.GetCustomerByEmail(email)
}

// HandleWebhook processes webhook events
func (s *PaymentService) HandleWebhook(event *models.PaystackWebhookEvent, payload string) error {
	// Store webhook
	webhook := &models.Webhook{
		EventType: event.Event,
		Payload:   payload,
		Processed: false,
	}

	if err := s.repo.CreateWebhook(webhook); err != nil {
		return fmt.Errorf("failed to store webhook: %w", err)
	}

	// Process based on event type
	switch event.Event {
	case "charge.success":
		return s.handleChargeSuccess(event, webhook.ID)
	case "charge.failed":
		return s.handleChargeFailed(event, webhook.ID)
	default:
		s.logger.Infof("Unhandled webhook event: %s", event.Event)
	}

	return s.repo.MarkWebhookProcessed(webhook.ID)
}

func (s *PaymentService) handleChargeSuccess(event *models.PaystackWebhookEvent, webhookID uint) error {
	transaction, err := s.repo.GetTransactionByReference(event.Data.Reference)
	if err != nil {
		s.logger.Errorf("Transaction not found for webhook: %s", event.Data.Reference)
		return err
	}

	transaction.Status = models.StatusSuccess
	transaction.PaystackRef = event.Data.Reference

	if err := s.repo.UpdateTransaction(transaction); err != nil {
		return err
	}

	s.logger.Infof("Webhook processed: charge.success for %s", event.Data.Reference)
	return s.repo.MarkWebhookProcessed(webhookID)
}

func (s *PaymentService) handleChargeFailed(event *models.PaystackWebhookEvent, webhookID uint) error {
	transaction, err := s.repo.GetTransactionByReference(event.Data.Reference)
	if err != nil {
		s.logger.Errorf("Transaction not found for webhook: %s", event.Data.Reference)
		return err
	}

	transaction.Status = models.StatusFailed

	if err := s.repo.UpdateTransaction(transaction); err != nil {
		return err
	}

	s.logger.Infof("Webhook processed: charge.failed for %s", event.Data.Reference)
	return s.repo.MarkWebhookProcessed(webhookID)
}

func (s *PaymentService) mapPaystackStatus(status string) models.TransactionStatus {
	switch status {
	case "success":
		return models.StatusSuccess
	case "failed":
		return models.StatusFailed
	case "abandoned":
		return models.StatusAbandoned
	default:
		return models.StatusPending
	}
}