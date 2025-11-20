package ratelimit

import "context"

type Limiter interface {
	Allow(ctx context.Context, tenantID, key string, tokens int) (bool, error)
}
