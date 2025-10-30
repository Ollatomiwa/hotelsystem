package retry

import (

	"math/rand"
	"strings"
	"time"
)

// RetryConfig holds retry configuration
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() *RetryConfig {
    return &RetryConfig{
        MaxAttempts: 3,
        InitialDelay: 100 * time.Millisecond,
        MaxDelay:     5 * time.Second,
        Multiplier:   2.0,
    }
}

// Retry executes a function with retry logic
func Retry(config *RetryConfig, fn func() error) error {
    var err error
    delay := config.InitialDelay

    for attempt := 0; attempt < config.MaxAttempts; attempt++ {
        err = fn()
        if err == nil {
            return nil // Success
        }

        // Don't retry if it's the last attempt
        if attempt == config.MaxAttempts-1 {
            break
        }

        // Calculate next delay with jitter
        nextDelay := time.Duration(float64(delay) * config.Multiplier)
        if nextDelay > config.MaxDelay {
            nextDelay = config.MaxDelay
        }

        // Add jitter to prevent thundering herd
        jitter := time.Duration(rand.Int63n(int64(nextDelay) / 4))
        if rand.Float64() < 0.5 {
            jitter = -jitter
        }

        sleepTime := nextDelay + jitter
        time.Sleep(sleepTime)
        delay = nextDelay
    }

    return err
}

// IsRetryableError checks if an error is worth retrying
func IsRetryableError(err error) bool {
    if err == nil {
        return false
    }

    errorMsg := err.Error()
    
    // Retry on temporary network issues
    retryableMessages := []string{
        "timeout",
        "connection refused",
        "network",
        "temporary",
        "busy",
        "overloaded",
    }

    for _, msg := range retryableMessages {
        if strings.Contains(strings.ToLower(errorMsg), msg) {
            return true
        }
    }

    return false
}
