package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"github.com/ollatomiwa/hotelsystem/payment-service/config"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/repository"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/router"
	"github.com/ollatomiwa/hotelsystem/payment-service/internal/service"
	"github.com/ollatomiwa/hotelsystem/payment-service/pkg/database"
	"github.com/ollatomiwa/hotelsystem/payment-service/pkg/paystack"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.IsProduction() {
		logger.SetLevel(logrus.WarnLevel)
	} else {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Info("Starting payment service...")

	// Ensure data directory exists
	if err := config.EnsureDataDirectory(cfg.Database.Path); err != nil {
		logger.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize database
	db, err := database.New(cfg.Database.Path, cfg.IsProduction())
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	logger.Info("Running database migrations...")
	if err := db.AutoMigrate(); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repository
	repo := repository.New(db.DB)

	// Initialize Paystack client
	paystackClient := paystack.NewClient(
		cfg.Paystack.SecretKey,
		cfg.Paystack.BaseURL,
		logger,
	)

	// Initialize service
	paymentService := service.NewPaymentService(repo, paystackClient, logger)

	// Initialize handlers
	paymentHandler := handlers.NewPaymentHandler(paymentService, logger, cfg.Paystack.SecretKey)
	healthHandler := handlers.NewHealthHandler(db)

	// Setup router
	r := router.Setup(cfg, paymentHandler, healthHandler, logger)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("Server listening on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}