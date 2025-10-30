package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/ollatomiwa/hotelsystem/notification-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories/postgres"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/config"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/email"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/health"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/logging"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/middleware"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/ratelimiter"
)

// initDB initializes the PostgreSQL database and creates tables
func initDB(connectionString string) (*sql.DB, error) {
    if connectionString == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }

    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Create notifications table for PostgreSQL
    createTableSQL := `
        CREATE TABLE IF NOT EXISTS notifications (
            id TEXT PRIMARY KEY,
            to_email TEXT NOT NULL,
            subject TEXT NOT NULL,
            body TEXT NOT NULL,
            status TEXT NOT NULL,
            type TEXT NOT NULL,
            retry_count INTEGER DEFAULT 0,
            sent_at TEXT NULL,
            error TEXT DEFAULT ''
        )
    `

    _, err = db.Exec(createTableSQL)
    if err != nil {
        return nil, err
    }

    log.Println("PostgreSQL database initialized successfully")
    return db, nil
}
//setupROUTER to initialize all depencies and set up http routes
func setupRouter(cfg *config.Config) (*gin.Engine, error) {
	//initializes db
	db, err := initDB(cfg.DatabaseURL)
	if err != nil {
		return nil, err 
	}

	//initialize logging
	logger := logging.NewLogger("notification-service")

	//initialize health checker
	healthChecker := health.NewHealthChecker()
    healthChecker.RegisterCheck("database", health.DatabaseCheck(db))
    healthChecker.RegisterCheck("smtp", health.SMTPCheck(
        cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword,
    ))

	//initializes repos
	notificationRepo := postgres.NewNotificationRepo(db)

	//initialize email sender
	emailSender := email.NewSMTPSender(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUsername,
		cfg.SMTPPassword,
		cfg.FromeEmail,
	)

	//initialize rate limiter
	rateLimiter := ratelimiter.NewRateLimiter(
		cfg.RateLimitRequest,
		cfg.RateLimitMinutes,
	)

	//testing email config in development
	if cfg.Environment == "development" {
		log.Println("testing SMTP conn..")
		if err := emailSender.TestConnection(); err != nil {
			log.Printf("warning: SMTP conn test failed: %v", err)
			log.Println("emails may not send properly, check your smtp configuration")
		} else {
			log.Println("SMTP conn test successful!")
		}
	}

	//initializes service
	notificationService := services.NewNotificationService(
		notificationRepo,
		emailSender,
		rateLimiter,
	)
	//initilizes handlers
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	//setup gin router : use new instead
	router:= gin.New()

	//add security middleware
	securityConfig := &middleware.SecurityConfig{
		MaxBodySize: int(cfg.MaxRequestBodySize),
		AllowedOrigins: cfg.AllowedOrigins,
	}

	router.Use(middleware.SecurityMiddleware(securityConfig))

	//add middleware
	router.Use(logger.RequestIDMiddleware())
	router.Use(logger.LoggingMiddleware())
	router.Use(middleware.SecurityMiddleware(securityConfig))
	router.Use(gin.Recovery())

	// Enhanced health check with detailed status
    router.GET("/health", func(c *gin.Context) {
        status := healthChecker.Check(c.Request.Context())
        c.JSON(200, status)
    })
	 router.GET("/ready", func(c *gin.Context) {
    // Simple readiness check - just database
    if err := db.PingContext(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{"status": "not ready", "error": err.Error()})
        	 return
        }
        c.JSON(200, gin.H{"status": "ready"})
    })

	//API routes
	api := router.Group("/api/v1")
	{
		notifications := api.Group("/notifications")
		{
			notifications.POST("/email", notificationHandler.SendEmail)
			notifications.GET("/:id", notificationHandler.GetNotificationStatus)
		}
	}

	 log.Printf("Monitoring: Structured logging enabled")
    log.Printf("Monitoring: Health checks registered for database and SMTP")
	return router, nil
}

func main () {
	cfg := config.Load()
	//setup router
	router, err := setupRouter(cfg)
	if err != nil {
		log.Fatalf("Failed to setup router:%v", err)
	}

	//creating HTTP server
	server := &http.Server{
		Addr: ":" + cfg.ServerPort,
		Handler: router,
		ReadTimeout: 15 *time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout: 60 *time.Second,
	}

	go func() {
		//starting server in a goroutine
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start:%v", err)
		}
	}()
	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
	
}