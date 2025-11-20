package ratelimit

import "context"

// NoopLimiter is a rate limiter implementation that always allows requests.
// It should only be used for development or testing purposes.
type NoopLimiter struct{}

// NewNoopLimiter creates a new NoopLimiter instance.
func NewNoopLimiter() *NoopLimiter {
	return &NoopLimiter{}
}

// Allow always returns true and does not enforce any limits.
func (l *NoopLimiter) Allow(ctx context.Context, tenantID, key string, tokens int) (bool, error) {
	return true, nil
}
