package httpx

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
)

func SetRateLimitHeaders(h http.Header, r ratelimit.Result, now time.Time) {
	if r.Limit > 0 {
		h.Set("X-Ratelimit-Limit", strconv.Itoa(r.Limit))
	}

	if r.Remaining > 0 {
		h.Set("X-Ratelimit-Remaining", strconv.Itoa(r.Remaining))
	}

	if r.ResetAfter > 0 {
		// seconds until reset
		h.Set("X-Ratelimit-Reset", strconv.FormatInt(int64(r.ResetAfter.Seconds()), 10))
	}

	// Retry-After just if blocked
	if !r.Allowed && r.RetryAfter > 0 {
		// seconds form is the simplest to client retry
		h.Set("Retry-After", strconv.FormatInt(int64(r.RetryAfter.Seconds()), 10))
	}
}
