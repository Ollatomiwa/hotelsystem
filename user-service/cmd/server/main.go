	package main

import (
	"log"

	"github.com/ollatomiwa/hotelsystem/user-service/internal/handlers"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/config"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/database"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/security"

	"github.com/ollatomiwa/hotelsystem/user-service/internal/repositories/postgres"
	"github.com/ollatomiwa/hotelsystem/user-service/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg :=config.Load()

	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}else {
		gin.SetMode(gin.DebugMode)
	}

	db, err := database.NewPostgresConn(database.Config{
		Host: cfg.Database.Host,
		Port: cfg.Database.Port,
		User: cfg.Database.User,
		Password: cfg.Database.Password,
		DBName: cfg.Database.DBName,
		SSLMode: cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := database.InitializeSchema(db); err != nil {
		log.Fatal("failed to initialize database schema:",err)
	}

	jwtManager := security.NewJWTManager(
		cfg.Security.JWTSecretKey,
		cfg.Security.JWTRefreshKey,
		cfg.Security.AccessTokenDuration,
		cfg.Security.RefreshToken,	
	)

	userRepo := postgres.NewUserRepository(db)
	authService := services.NewAuthService(userRepo, jwtManager)

	router := gin.Default()

	handlers.SetupRoutes(router, authService, jwtManager)

	log.Printf("user service starting on port %s", cfg.Server.Port)
	log.Printf("Environment: %s", cfg.Server.Env)
	log.Printf("JWT Acess DUration: %v", cfg.Security.AccessTokenDuration)
	log.Printf("JWT Refresh Duration: %v", cfg.Security.RefreshToken)

	address := ":" + cfg.Server.Port
	if err := router.Run(address); err != nil {
		log.Fatal("failed to start server:", err)
	}
}