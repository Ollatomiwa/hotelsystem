package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//security config holds security configs
type SecurityConfig struct {
	MaxBodySize int
	AllowedOrigins string 
}

//defualt security config returns a sensible default security config
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		MaxBodySize: 1 * 1024 *1024, //1mb,
		AllowedOrigins: "*", //restrict in prod
	}
}

//security middleware provides security prod
func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		//setting security header
		c.Header("X-Content-Type-Optiond", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		//limit request body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(config.MaxBodySize))

		//basic CORS handling 
		origin := c.Request.Header.Get("Origin")
		if config.AllowedOrigins == "*" || origin == config.AllowedOrigins {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		//handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return 
		}
		c.Next()
	}
}