package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ollatomiwa/hotelsystem/notification-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/repositories/sqlite"
	"github.com/ollatomiwa/hotelsystem/notification-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/config"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/email"
	"github.com/ollatomiwa/hotelsystem/notification-service/pkg/ratelimiter"
	
)

//initDB initializes sqlited db and creates tables
func initDB(dbPath string) (*sql.DB, error){
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err 
	}

	//creating notification tables
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS notifications (
		id TEXT PRIMARY KEY,
		to_email TEXT NOT NULL,
		subject TEXT NOT NULL,
		body TEXT NOT NULL,
		status TEXT NOT NULL,
		type TEXT NOT NULL,
		retry_count INTEGER DEFAULT 0,
		sent_at TEXT NOT NULL,
		error TEXT DEFAULT ''
		)
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err 
	}

	//Enable WAL Mode for bettter concurrency
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, err 
	}

	log.Println("Database initialized successfully")
	return db, nil

}
//setupROUTER to initialize all depencies and set up http routes
func setupRouter(cfg *config.Config) (*gin.Engine, error) {
	//initializes db
	db, err := initDB(cfg.DatabasePath)
	if err != nil {
		return nil, err 
	}

	//initializes repos
	notificationRepo := sqlite.NewNotificationRepo(db)

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

	//setuo gin router
	router:= gin.Default()

	//add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	//healthcheck
	router.GET("/health", notificationHandler.HealthCheck)

	//API routes
	api := router.Group("/api/v1")
	{
		notifications := api.Group("/notifications")
		{
			notifications.POST("/email", notificationHandler.SendEmail)
			notifications.GET("/:id", notificationHandler.GetNotificationStatus)
		}
	}

	log.Printf("rate limiting: %d requests per %d minutes", cfg.RateLimitRequest, cfg.RateLimitMinutes)
	log.Printf("SMTP configured: %s:%d", cfg.SMTPHost, cfg.SMTPPort)
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