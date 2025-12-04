package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RequestLogger logs HTTP requests
func RequestLogger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.WithFields(logrus.Fields{
			"method":     method,
			"path":       path,
			"status":     statusCode,
			"duration":   duration,
			"request_id": c.GetString("request_id"),
			"ip":         c.ClientIP(),
		}).Info("HTTP request")
	}
}

// RequestID adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// APIKeyAuth validates API key authentication
func APIKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for health check and webhook endpoints
		if c.Request.URL.Path == "/api/v1/health" || c.Request.URL.Path == "/api/v1/webhooks/paystack" {
			c.Next()
			return
		}

		key := c.GetHeader("X-API-Key")
		if key == "" {
			key = c.GetHeader("Authorization")
			if len(key) > 7 && key[:7] == "Bearer " {
				key = key[7:]
			}
		}

		if key != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status":  "error",
				"message": "Invalid or missing API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiter implements rate limiting per IP
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.limiters[ip] = limiter
	}

	return limiter
}

func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"status":  "error",
				"message": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Recovery recovers from panics
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"error":      err,
					"request_id": c.GetString("request_id"),
				}).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// IdempotencyKey extracts and validates idempotency key
func IdempotencyKey() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			key := c.GetHeader("Idempotency-Key")
			if key != "" {
				c.Set("idempotency_key", key)
			}
		}
		c.Next()
	}
}