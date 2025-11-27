package ratelimit

import "errors"

type ErrRateLimited struct {
	result Result
	err    error
}

// Error implements error.
func (e *ErrRateLimited) Error() string {
	return e.err.Error()
}

func (r Result) AsError() error {
	return &ErrRateLimited{
		result: r,
		err:    errors.New("rate limite reached"),
	}
}

func (e *ErrRateLimited) Result() Result {
	return e.result
}
