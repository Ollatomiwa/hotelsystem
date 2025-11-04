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
	cfg :=config.Load()

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := database.NewPostgresConn(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatal("failed to connect to database", err)
	}
	defer db.Close()

	if err := database.InitializeSchema(db); err != nil {
		log.Fatal("failed to initialize database schema:", err)
	}

	bookingRepo := postgres.NewBookingRepository(db)
	roomRepo := postgres.NewRoomRepository(db)

	bookingService := services.NewBookingService(bookingRepo, roomRepo)

	router := gin.Default()

	handlers.SetupRoutes(router, bookingService)

	log.Printf("BOoking service is starting on port %s", cfg.Server.Port)
	log.Printf("Environment %s", cfg.Server.Env)

	if err := router.Run(":", cfg.Server.Port); err != nil {
		log.Fatal("failed to start server:", err)
	}
}