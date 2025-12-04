package router

import (
	"github.com/ollatomiwa/hotelsystem/payment-service/config"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Setup(
	cfg *config.Config,
	paymentHandler *handlers.PaymentHandler,
	healthHandler *handlers.HealthHandler,
	logger *logrus.Logger,
) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// CORS configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"X-API-Key",
		"X-Request-ID",
		"Idempotency-Key",
	}
	r.Use(cors.New(corsConfig))

	// Global middlewares
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.IdempotencyKey())

	// Rate limiting
	rateLimiter := middleware.NewRateLimiter(
		cfg.RateLimit.Requests/cfg.RateLimit.Window,
		cfg.RateLimit.Requests,
	)
	r.Use(rateLimiter.Middleware())

	// API v1 group
	v1 := r.Group("/api/v1")

	// Health check (no auth required)
	v1.GET("/health", healthHandler.HealthCheck)

	// Webhook endpoint (no auth, uses signature verification)
	v1.POST("/webhooks/paystack", paymentHandler.HandleWebhook)

	// Protected routes
	if cfg.Security.APIKey != "" {
		v1.Use(middleware.APIKeyAuth(cfg.Security.APIKey))
	}

	// Payment routes
	payments := v1.Group("/payments")
	{
		payments.POST("/initialize", paymentHandler.InitializePayment)
		payments.GET("/verify/:reference", paymentHandler.VerifyPayment)
		payments.GET("/:id", paymentHandler.GetPayment)
		payments.GET("", paymentHandler.ListPayments)
	}

	// Customer routes
	customers := v1.Group("/customers")
	{
		customers.GET("/:email", paymentHandler.GetCustomer)
	}

	return r
}