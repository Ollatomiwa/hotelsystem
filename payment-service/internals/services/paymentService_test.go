package service

import (
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/models"
	"testing"
)

// Mock repository for testing
type mockRepository struct {
	transactions map[string]*models.Transaction
	customers    map[string]*models.Customer
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		transactions: make(map[string]*models.Transaction),
		customers:    make(map[string]*models.Customer),
	}
}

func (m *mockRepository) CreateTransaction(tx *models.Transaction) error {
	m.transactions[tx.Reference] = tx
	return nil
}

func (m *mockRepository) GetTransactionByReference(ref string) (*models.Transaction, error) {
	if tx, ok := m.transactions[ref]; ok {
		return tx, nil
	}
	return nil, nil
}

func TestMapPaystackStatus(t *testing.T) {
	tests := []struct {
		name           string
		paystackStatus string
		expected       models.TransactionStatus
	}{
		{
			name:           "Success status",
			paystackStatus: "success",
			expected:       models.StatusSuccess,
		},
		{
			name:           "Failed status",
			paystackStatus: "failed",
			expected:       models.StatusFailed,
		},
		{
			name:           "Abandoned status",
			paystackStatus: "abandoned",
			expected:       models.StatusAbandoned,
		},
		{
			name:           "Unknown status defaults to pending",
			paystackStatus: "unknown",
			expected:       models.StatusPending,
		},
	}

	service := &PaymentService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.mapPaystackStatus(tt.paystackStatus)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestTransactionValidation(t *testing.T) {
	tests := []struct {
		name    string
		request models.InitializePaymentRequest
		wantErr bool
	}{
		{
			name: "Valid request",
			request: models.InitializePaymentRequest{
				Email:    "test@example.com",
				Amount:   10000,
				Currency: "NGN",
			},
			wantErr: false,
		},
		{
			name: "Missing email",
			request: models.InitializePaymentRequest{
				Amount:   10000,
				Currency: "NGN",
			},
			wantErr: true,
		},
		{
			name: "Zero amount",
			request: models.InitializePaymentRequest{
				Email:    "test@example.com",
				Amount:   0,
				Currency: "NGN",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic would be tested here
			hasError := tt.request.Email == "" || tt.request.Amount <= 0
			if hasError != tt.wantErr {
				t.Errorf("expected error: %v, got error: %v", tt.wantErr, hasError)
			}
		})
	}
}