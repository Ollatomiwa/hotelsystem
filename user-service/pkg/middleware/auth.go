package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ollatomiwa/hotelsystem/user-service/pkg/security"
)

//vaidates JWT tokens and set user context
func AuthMiddleware(jwtManager *security.JWTManager) (gin.HandlerFunc) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return 
		}

		//extract token from bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return 
		}

		tokenString :=parts[1]

		//verify token
		claims, err := jwtManager.VerifyAccessToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return 
		}

		//set user info in context
		c.Set("userId", claims.UserId)
		c.Set("userRole", claims.Role)
		c.Set("userEmail", claims.UserId)//temporary - replace later with actual email lookup

		c.Next()
	
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user role not found"})
			c.Abort()
			return 
		}

		if userRole != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error":"insufficient permissions"})
			c.Abort()
			return 
		}
		c.Next()
	}
}

// OptionalAuthMiddleware validates JWT tokens if present, but doesn't require them
func OptionalAuthMiddleware(jwtManager *security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]

		// Verify token
		claims, err := jwtManager.VerifyAccessToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Set user information in context
		c.Set("userID", claims.UserId)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}