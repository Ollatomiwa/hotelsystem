package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
    StateClosed State = iota    // Normal operation
    StateOpen                  // Blocking all requests
    StateHalfOpen              // Testing if service recovered
)

// CircuitBreaker protects against cascading failures
type CircuitBreaker struct {
    state         State
    failures      int
    successCount  int
    lastFailure   time.Time
    mu            sync.RWMutex
    
    // Configuration
    failureThreshold int
    successThreshold int
    timeout          time.Duration
    halfOpenMax      int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold, successThreshold int, timeout time.Duration) *CircuitBreaker {
    return &CircuitBreaker{
        state:            StateClosed,
        failureThreshold: failureThreshold,
        successThreshold: successThreshold,
        timeout:          timeout,
        halfOpenMax:      1, // Only allow 1 request in half-open state
    }
}

// Execute runs a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
    cb.mu.RLock()
    state := cb.state
    cb.mu.RUnlock()

    // Check if circuit is open
    if state == StateOpen {
        if time.Since(cb.lastFailure) > cb.timeout {
            // Timeout reached, try half-open state
            cb.mu.Lock()
            cb.state = StateHalfOpen
            cb.successCount = 0
            cb.mu.Unlock()
        } else {
            return ErrCircuitOpen
        }
    }

    // Execute the function
    err := fn()

    cb.mu.Lock()
    defer cb.mu.Unlock()

    if err != nil {
        // Request failed
        cb.failures++
        cb.lastFailure = time.Now()

        if cb.state == StateHalfOpen || cb.failures >= cb.failureThreshold {
            cb.state = StateOpen
            cb.failures = 0
        }
        return err
    }

    // Request succeeded
    if cb.state == StateHalfOpen {
        cb.successCount++
        if cb.successCount >= cb.successThreshold {
            cb.state = StateClosed
            cb.failures = 0
            cb.successCount = 0
        }
    } else {
        cb.failures = 0 // Reset on success in closed state
    }

    return nil
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() State {
    cb.mu.RLock()
    defer cb.mu.RUnlock()
    return cb.state
}

// Errors
var (
    ErrCircuitOpen = errors.New("circuit breaker is open")
)

func init() {
    errors.New("circuit breaker is open")
}