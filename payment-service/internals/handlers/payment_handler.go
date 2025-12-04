package handlers

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"github.com/ollatomiwa/hotelsystem/payment-service/internals/models"
	"github.com/ollatomiwa/hotelsystem/payment-service/internals/service"
	"github.com/ollatomiwa/hotelsystem/payment-service/pkg/paystack"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	service        *service.PaymentService
	logger         *logrus.Logger
	paystackSecret string
}

func NewPaymentHandler(service *service.PaymentService, logger *logrus.Logger, paystackSecret string) *PaymentHandler {
	return &PaymentHandler{
		service:        service,
		logger:         logger,
		paystackSecret: paystackSecret,
	}
}

// InitializePayment godoc
// @Summary Initialize payment
// @Description Initialize a new payment transaction
// @Tags payments
// @Accept json
// @Produce json
// @Param request body models.InitializePaymentRequest true "Payment request"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/payments/initialize [post]
func (h *PaymentHandler) InitializePayment(c *gin.Context) {
	var req models.InitializePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// Get idempotency key if present
	if key, exists := c.Get("idempotency_key"); exists {
		req.IdempotencyKey = key.(string)
	}

	transaction, err := h.service.InitializePayment(&req)
	if err != nil {
		h.logger.Errorf("Failed to initialize payment: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to initialize payment",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Payment initialized successfully",
		Data:    transaction,
	})
}

// VerifyPayment godoc
// @Summary Verify payment
// @Description Verify a payment transaction
// @Tags payments
// @Produce json
// @Param reference path string true "Transaction reference"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/payments/verify/{reference} [get]
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	reference := c.Param("reference")
	if reference == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Reference is required",
		})
		return
	}

	transaction, err := h.service.VerifyPayment(reference)
	if err != nil {
		h.logger.Errorf("Failed to verify payment: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to verify payment",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Payment verified successfully",
		Data:    transaction,
	})
}

// GetPayment godoc
// @Summary Get payment details
// @Description Get details of a specific payment
// @Tags payments
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/payments/{id} [get]
func (h *PaymentHandler) GetPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Invalid transaction ID",
		})
		return
	}

	transaction, err := h.service.GetTransaction(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Status:  "error",
				Message: "Transaction not found",
			})
			return
		}
		h.logger.Errorf("Failed to get payment: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to retrieve payment",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Payment retrieved successfully",
		Data:    transaction,
	})
}

// ListPayments godoc
// @Summary List payments
// @Description List all payments with pagination
// @Tags payments
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "Filter by status"
// @Param email query string false "Filter by customer email"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/payments [get]
func (h *PaymentHandler) ListPayments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	email := c.Query("email")

	transactions, total, err := h.service.ListTransactions(page, pageSize, status, email)
	if err != nil {
		h.logger.Errorf("Failed to list payments: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to retrieve payments",
		})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Payments retrieved successfully",
		Data:    transactions,
		Meta: &models.Meta{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetCustomer godoc
// @Summary Get customer details
// @Description Get details of a specific customer
// @Tags customers
// @Produce json
// @Param email path string true "Customer email"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/customers/{email} [get]
func (h *PaymentHandler) GetCustomer(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Email is required",
		})
		return
	}

	customer, err := h.service.GetCustomer(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, models.APIResponse{
				Status:  "error",
				Message: "Customer not found",
			})
			return
		}
		h.logger.Errorf("Failed to get customer: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to retrieve customer",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Customer retrieved successfully",
		Data:    customer,
	})
}

// HandleWebhook godoc
// @Summary Handle Paystack webhook
// @Description Handle webhook events from Paystack
// @Tags webhooks
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse
// @Router /api/v1/webhooks/paystack [post]
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Paystack-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Missing signature",
		})
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Failed to read request body",
		})
		return
	}

	// Validate signature
	if !paystack.ValidateWebhookSignature(signature, body, h.paystackSecret) {
		h.logger.Warn("Invalid webhook signature")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Status:  "error",
			Message: "Invalid signature",
		})
		return
	}

	var event models.PaystackWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Status:  "error",
			Message: "Invalid webhook payload",
		})
		return
	}

	if err := h.service.HandleWebhook(&event, string(body)); err != nil {
		h.logger.Errorf("Failed to handle webhook: %v", err)
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Status:  "error",
			Message: "Failed to process webhook",
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: "Webhook processed successfully",
	})
}