package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/user-service/internal/services"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/middleware"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/security"
)

func SetupRoutes(router *gin.Engine, authService *services.AuthService, jwtManager *security.JWTManager){
	AuthHandler := NewAuthHandler(authService)

	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	router.GET("/health", func (c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message":"healthy"})
	})

	v1 := router.Group("/api/v1")
	{
		//auth routes - public
		auth := v1.Group("/auth")
		{
			auth.POST("/reigister", AuthHandler.Register)
			auth.POST("/login", AuthHandler.Login)
			auth.POST("/refresh", AuthHandler.RefreshToken)
		}

		//user routes- protected
		users := v1.Group("/users")	
		users.Use(middleware.AuthMiddleware(jwtManager))
		{
			users.GET("/profile", AuthHandler.GetProfile)
			users.PUT("/profile", AuthHandler.UpdateProfile)
			users.PUT("/change-password", AuthHandler.ChangePassword)
		}

		//admin routes - protected + admin role
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware(jwtManager))
		admin.Use(middleware.RoleMiddleware("admin"))
		{
			admin.Group("/Users", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Admin access granted"})
			})
		}
	}

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})
}