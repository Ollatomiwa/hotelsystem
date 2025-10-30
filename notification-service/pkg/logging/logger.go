package logging

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
    RequestIDKey contextKey = "requestID"
)

// StructuredLog represents a structured log entry
type StructuredLog struct {
    Timestamp  string                 `json:"timestamp"`
    Level      string                 `json:"level"`
    Message    string                 `json:"message"`
    RequestID  string                 `json:"request_id,omitempty"`
    Service    string                 `json:"service"`
    Duration   int64                 `json:"duration_ms,omitempty"`
    Additional map[string]interface{} `json:"additional,omitempty"`
}

// Logger handles structured logging
type Logger struct {
    serviceName string
}

// NewLogger creates a new logger instance
func NewLogger(serviceName string) *Logger {
    return &Logger{
        serviceName: serviceName,
    }
}

// RequestIDMiddleware adds a request ID to each request
func (l *Logger) RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.Request.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set(string(RequestIDKey), requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// LoggingMiddleware logs each request with structured data
func (l *Logger) LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // Process request
        c.Next()

        duration := time.Since(start)
        
        // Create structured log
        logEntry := StructuredLog{
            Timestamp: time.Now().Format(time.RFC3339),
            Level:     "info",
            Message:   "request completed",
            RequestID: c.GetString(string(RequestIDKey)),
            Service:   l.serviceName,
            Duration:  duration.Milliseconds(),
            Additional: map[string]interface{}{
                "method":     c.Request.Method,
                "path":       c.Request.URL.Path,
                "status":     c.Writer.Status(),
                "client_ip":  c.ClientIP(),
                "user_agent": c.Request.UserAgent(),
            },
        }

        l.Info(logEntry)
    }
}

// Info logs an info level message
func (l *Logger) Info(logEntry StructuredLog) {
    logEntry.Level = "info"
    l.printLog(logEntry)
}

// Error logs an error level message
func (l *Logger) Error(logEntry StructuredLog) {
    logEntry.Level = "error"
    l.printLog(logEntry)
}

// Warn logs a warning level message
func (l *Logger) Warn(logEntry StructuredLog) {
    logEntry.Level = "warn"
    l.printLog(logEntry)
}

// printLog outputs the structured log (in production, this would send to a logging service)
func (l *Logger) printLog(logEntry StructuredLog) {
    // For now, we'll print as JSON. In production, send to ELK/Splunk/etc.
    log.Printf("STRUCTURED_LOG: %+v", logEntry)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
    if ginCtx, ok := ctx.(*gin.Context); ok {
        return ginCtx.GetString(string(RequestIDKey))
    }
    
    if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
        return reqID
    }
    
    return ""
}
