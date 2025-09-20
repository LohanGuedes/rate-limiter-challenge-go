package ratelimit

import (
	"fmt"
	"time"
)

// LimitExceededError represents a rate limit exceeded error with retry timing
type LimitExceededError struct {
	RetryAfter time.Duration
	Message    string
}

func (e *LimitExceededError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("rate limit exceeded, retry after %v", e.RetryAfter)
}

// NewLimitExceededError creates a new rate limit exceeded error
func NewLimitExceededError(retryAfter time.Duration, message string) *LimitExceededError {
	return &LimitExceededError{
		RetryAfter: retryAfter,
		Message:    message,
	}
}
