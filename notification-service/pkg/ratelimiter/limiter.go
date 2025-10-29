package ratelimiter

import (
	"sync"
	"time"
)

type RateLimiter struct {
	limits map[string][]time.Time
	mu sync.RWMutex
	request int
	window time.Duration
}

//create newratelimiter to create a new rate limiter
func NewRateLimiter( request int, minutes int) *RateLimiter {
	return &RateLimiter{
		limits: make(map[string][]time.Time),
		request: request,
		window: time.Duration(minutes) *time.Minute,
	}
}

//allow checks if a req is allowed for an identifier
func (r *RateLimiter) Allow(identifier string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now:= time.Now()
	windowsStart := now.Add(-r.window)

	//getting existing reqs for this identifier
	requests, exists := r.limits[identifier]
	if !exists {
		requests = []time.Time{}
	}

	//remove old reqs outside the time window
	var validRequests []time.Time
	for _, t := range requests {
		if t.After(windowsStart) {
			validRequests = append(validRequests, t)
		}
	}

	//check if under the limit
	if len(validRequests) >= r.request {
		return false 
	}

	//append current reqs and updates
	validRequests = append(validRequests, now)
	r.limits[identifier] = validRequests
	return  true
}

//getremainingreqs returns how many reqs are left for an indentifier
func (r *RateLimiter) GetRemainingRequests(identifier string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	windowsStart := now.Add(-r.window)

	requests, exists := r.limits[identifier]
	if !exists {
		return r.request
	}

	//counts valid request within time window
	validCount := 0
	for _, t := range requests {
		if t.After(windowsStart) {
			validCount++
		}
	}

	remaining := r.request - validCount
	if remaining < 0 {
		return 0
	}
	return remaining
}

