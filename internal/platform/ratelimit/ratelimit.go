package ratelimit

import (
	"context"
	"errors"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/config"
)

// Limiter defines the interface for rate limiters.
//
// It is tenant-aware and allows checking if a given operation identified by
// "key" is allowed to consume the given number of "tokens".
type Limiter interface {
	Allow(ctx context.Context, tenantID, key string, tokens int) (bool, error)
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
