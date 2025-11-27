package ratelimit

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
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

// allow processes the request and returns:
//
//	allowed, remaining, retryAfter, resetAfter.
func (b *tokenBucket) allow(now time.Time, requested int) (bool, int, time.Duration, time.Duration) {
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

	// When is not enough tokens, calc Retry-After
	if b.tokens < float64(requested) {
		deficit := float64(requested) - b.tokens

		var retryAfter time.Duration

		if b.refillRate > 0 {
			sec := deficit / b.refillRate
			retryAfter = time.Duration(math.Ceil(sec)) * time.Second
		}

		// How much time until the bucket resets
		var resetAfter time.Duration

		if b.refillRate > 0 {
			defFul := b.maxTokens - b.tokens
			sec := defFul / b.refillRate
			resetAfter = time.Duration(math.Ceil(sec)) * time.Second
		}

		return false, int(b.tokens), retryAfter, resetAfter
	}

	// consume tokens
	b.tokens -= float64(requested)

	var resetAfter time.Duration

	if b.refillRate > 0 {
		defFul := b.maxTokens - b.tokens
		sec := defFul / b.refillRate
		resetAfter = time.Duration(math.Ceil(sec)) * time.Second
	}

	return true, int(b.tokens), 0, resetAfter
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
func (l *localLimiter) Allow(ctx context.Context, tenantID, key string, tokens int) (Result, error) {
	if ctx.Err() != nil {
		return Result{}, ctx.Err()
	}

	if tenantID == "" {
		return Result{}, errors.New("ratelimit: tenantID is required")
	}

	if key == "" {
		return Result{}, errors.New("ratelimit: key is required")
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

	allowed, remaining, retryAfter, resetAfter := bkt.allow(now, tokens)

	return Result{
		Allowed:    allowed,
		Limit:      actualMax,
		Remaining:  remaining,
		RetryAfter: retryAfter,
		ResetAfter: resetAfter,
	}, nil
}
