package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"log"
)

// Logger middleware
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		log.Printf("[%s] %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}