package app

import (
	"errors"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/apperr"
)

var ErrRateLimited = apperr.TooManyRequests(errors.New("rate_limited"))
