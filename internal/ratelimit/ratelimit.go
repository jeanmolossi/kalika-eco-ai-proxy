package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
)

// Result holds the decision and metadata for rate limiting.
type Result struct {
	// Allowed indicates whether the operation is permitted.
	Allowed bool

	// Limit is the total number of tokens allowed per window for this bucket.
	Limit int

	// Remaining is the estimated number of tokens remaining after this call.
	Remaining int

	// RetryAfter is the suggested duration the client should wait before retrying
	// the same operation. It is only meaningful when Allowed == false.
	RetryAfter time.Duration

	// ResetAfter is the estimated duration until the bucket is fully refilled.
	// Useful to populate "X-RateLimit-Reset" headers.
	ResetAfter time.Duration
}

// Limiter defines the interface for rate limiters.
//
// It is tenant-aware and allows checking if a given operation identified by
// "key" is allowed to consume the given number of "tokens".
type Limiter interface {
	Allow(ctx context.Context, tenantID, key string, tokens int) (Result, error)
}

// DefaultConfig returns a sane default configuration for local rate limiting.
func DefaultConfig() config.RateLimit {
	return config.RateLimit{
		Enabled:                  false,
		Backend:                  config.BackendNoop,
		DefaultTokensPerInterval: 60,
		Interval:                 time.Minute,
		MaxTokens:                0,
	}
}

// NewLimiter builds a Limiter from the given configuration.
//
// For now it supports:
//   - noop    -> no-op limiter that always allows
//   - local   -> in-memory token bucket, tenant-aware
func NewLimiter(cfg config.RateLimit) (Limiter, error) {
	if !cfg.Enabled || cfg.Backend == config.BackendNoop {
		return NewNoopLimiter(), nil
	}

	if cfg.Interval <= 0 {
		return nil, errors.New("ratelimit: interval must be > 0")
	}

	if cfg.DefaultTokensPerInterval <= 0 {
		return nil, errors.New("ratelimit: default tokens per interval must be > 0")
	}

	switch cfg.Backend {
	case config.BackendLocal:
		return newLocalLimiter(cfg), nil
	default:
		return nil, errors.New("ratelimit: unsupported backend: " + string(cfg.Backend))
	}
}

// bucketKey builds the internal key used to identify a rate limit bucket.
func bucketKey(tenantID, key string) string {
	// Simple concatenation is enough; we can change later if needed.
	return tenantID + ":" + key
}
