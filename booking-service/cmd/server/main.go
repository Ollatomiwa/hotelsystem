package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/repositories/postgres"
	"github.com/ollatomiwa/hotelsystem/booking-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/booking-service/pkg/config"
	"github.com/ollatomiwa/hotelsystem/booking-service/pkg/database"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database connection
	db, err := database.NewPostgresConn(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize database schema
	if err := database.InitializeSchema(db); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}

	// Initialize repositories
	bookingRepo := postgres.NewBookingRepository(db)
	roomRepo := postgres.NewRoomRepository(db)

	// Initialize services
	bookingService := services.NewBookingService(bookingRepo, roomRepo)

	// Create Gin router
	router := gin.Default()

	// Setup routes
	handlers.SetupRoutes(router, bookingService)

	// Start server - FIXED: Use proper port format
	address := ":" + cfg.Server.Port
	log.Printf("🚀 Booking Service starting on port %s", cfg.Server.Port)
	log.Printf("📊 Environment: %s", cfg.Server.Env)
	
	if err := router.Run(address); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}