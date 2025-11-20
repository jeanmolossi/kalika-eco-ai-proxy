package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/config"
)

// ----------------------------------------------------------------------------
// Local in-memory token bucket implementation
// ----------------------------------------------------------------------------

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

func newTokenBucket(now time.Time, maxTokens, tokensPerInterval float64, interval time.Duration) *tokenBucket {
	sec := interval.Seconds()
	if sec <= 0 {
		sec = 1
	}

	return &tokenBucket{
		tokens:     maxTokens, // start full
		maxTokens:  maxTokens,
		refillRate: tokensPerInterval / sec,
		lastRefill: now,
	}
}

func (b *tokenBucket) allow(now time.Time, requested int) bool {
	if requested <= 0 {
		requested = 1
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastRefill).Seconds()
	if elapsed > 0 {
		refilled := elapsed * b.refillRate
		if refilled > 0 {
			b.tokens += refilled
			if b.tokens > b.maxTokens {
				b.tokens = b.maxTokens
			}

			b.lastRefill = now
		}
	}

	if b.tokens < float64(requested) {
		return false
	}

	b.tokens -= float64(requested)

	return true
}

type localLimiter struct {
	cfg config.RateLimit
	now func() time.Time

	// buckets is keyed by "tenantID:key".
	buckets sync.Map // map[string]*tokenBucket
}

func newLocalLimiter(cfg config.RateLimit) *localLimiter {
	return &localLimiter{
		cfg:     cfg,
		now:     time.Now,
		buckets: sync.Map{},
	}
}

// Allow implements the Limiter interface for the local in-memory limiter.
//
// It uses a per-(tenantID,key) token bucket, which is safe for concurrent usage.
func (l *localLimiter) Allow(ctx context.Context, tenantID, key string, tokens int) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if tenantID == "" {
		return false, errors.New("ratelimit: tenantID is required")
	}

	if key == "" {
		return false, errors.New("ratelimit: key is required")
	}

	now := l.now()
	bKey := bucketKey(tenantID, key)

	actualMax := l.cfg.MaxTokens
	if actualMax <= 0 {
		actualMax = l.cfg.DefaultTokensPerInterval
	}

	val, _ := l.buckets.LoadOrStore(bKey, newTokenBucket(
		now,
		float64(actualMax),
		float64(l.cfg.DefaultTokensPerInterval),
		l.cfg.Interval,
	))

	bkt := val.(*tokenBucket)
	allowed := bkt.allow(now, tokens)

	return allowed, nil
}
