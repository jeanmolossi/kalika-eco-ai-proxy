package app

import (
	"errors"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/apperr"
)

var ErrRateLimited = apperr.TooManyRequests(errors.New("rate_limited"))
