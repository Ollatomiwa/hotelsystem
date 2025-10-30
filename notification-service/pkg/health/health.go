package health

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"sync"
	"time"
)

// HealthStatus represents the status of a service
type HealthStatus struct {
    Status    string            `json:"status"`
    Timestamp time.Time         `json:"timestamp"`
    Checks    map[string]Check  `json:"checks,omitempty"`
}

// Check represents an individual health check
type Check struct {
    Status    string    `json:"status"`
    Error     string    `json:"error,omitempty"`
    Timestamp time.Time `json:"timestamp"`
    Duration  int64     `json:"duration_ms"`
}

// HealthChecker manages health checks
type HealthChecker struct {
    checks map[string]CheckFunc
    mu     sync.RWMutex
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) error

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
    return &HealthChecker{
        checks: make(map[string]CheckFunc),
    }
}

// RegisterCheck registers a health check
func (h *HealthChecker) RegisterCheck(name string, check CheckFunc) {
    h.mu.Lock()
    defer h.mu.Unlock()
    h.checks[name] = check
}

// Check performs all registered health checks
func (h *HealthChecker) Check(ctx context.Context) HealthStatus {
    h.mu.RLock()
    defer h.mu.RUnlock()

    status := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]Check),
    }

    var wg sync.WaitGroup
    var mu sync.Mutex

    for name, checkFunc := range h.checks {
        wg.Add(1)
        
        go func(checkName string, fn CheckFunc) {
            defer wg.Done()
            
            start := time.Now()
            err := fn(ctx)
            duration := time.Since(start)
            
            mu.Lock()
            defer mu.Unlock()
            
            check := Check{
                Status:    "healthy",
                Timestamp: time.Now(),
                Duration:  duration.Milliseconds(),
            }
            
            if err != nil {
                check.Status = "unhealthy"
                check.Error = err.Error()
                status.Status = "unhealthy"
            }
            
            status.Checks[checkName] = check
        }(name, checkFunc)
    }
    
    wg.Wait()
    return status
}


// DatabaseCheck checks PostgreSQL database connectivity
func DatabaseCheck(db *sql.DB) CheckFunc {
    return func(ctx context.Context) error {
        return db.PingContext(ctx)
    }
}

// SMTPCheck checks SMTP connectivity
func SMTPCheck(host string, port int, username, password string) CheckFunc {
    return func(ctx context.Context) error {
        // Simple TCP connection test to SMTP server
        conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 5*time.Second)
        if err != nil {
            return fmt.Errorf("SMTP connection failed: %w", err)
        }
        defer conn.Close()
        return nil
    }
}