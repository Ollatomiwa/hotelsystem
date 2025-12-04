package repository

import (
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/models"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Transaction Repository Methods

func (r *Repository) CreateTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *Repository) GetTransactionByID(id uint) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *Repository) GetTransactionByReference(ref string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Where("reference = ?", ref).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *Repository) GetTransactionByIdempotencyKey(key string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Where("idempotency_key = ?", key).First(&tx).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *Repository) UpdateTransaction(tx *models.Transaction) error {
	return r.db.Save(tx).Error
}

func (r *Repository) ListTransactions(page, pageSize int, status string, email string) ([]models.Transaction, int64, error) {
	var transactions []models.Transaction
	var total int64

	query := r.db.Model(&models.Transaction{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if email != "" {
		query = query.Where("customer_email = ?", email)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&transactions).Error

	return transactions, total, err
}

// Customer Repository Methods

func (r *Repository) CreateCustomer(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *Repository) GetCustomerByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("email = ?", email).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *Repository) UpdateCustomer(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

func (r *Repository) GetOrCreateCustomer(email, name string) (*models.Customer, error) {
	customer, err := r.GetCustomerByEmail(email)
	if err == nil {
		return customer, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new customer
	customer = &models.Customer{
		Email: email,
		Name:  name,
	}
	if err := r.CreateCustomer(customer); err != nil {
		return nil, err
	}
	return customer, nil
}

// Webhook Repository Methods

func (r *Repository) CreateWebhook(webhook *models.Webhook) error {
	return r.db.Create(webhook).Error
}

func (r *Repository) GetWebhookByID(id uint) (*models.Webhook, error) {
	var webhook models.Webhook
	err := r.db.First(&webhook, id).Error
	if err != nil {
		return nil, err
	}
	return &webhook, nil
}

func (r *Repository) MarkWebhookProcessed(id uint) error {
	return r.db.Model(&models.Webhook{}).
		Where("id = ?", id).
		Update("processed", true).Error
}

func (r *Repository) GetUnprocessedWebhooks(limit int) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := r.db.Where("processed = ?", false).
		Order("created_at ASC").
		Limit(limit).
		Find(&webhooks).Error
	return webhooks, err
}