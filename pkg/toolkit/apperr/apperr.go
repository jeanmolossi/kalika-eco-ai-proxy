package apperr

import (
	"errors"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/ratelimit"
)

type Kind string

const (
	KindValidation   Kind = "validation_error"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindTooManyReq   Kind = "too_many_requests"
	KindInternal     Kind = "internal_error"
	KindBadGateway   Kind = "bad_gateway_error"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"` // e.g. required, email_invalid
}

type Error struct {
	kind       Kind
	httpStatus int
	title      string
	detail     string
	instance   string         // requrest-id / trace-id
	code       string         // client's stable business error
	fields     []FieldError   // field by field validation
	meta       map[string]any // debug safe for frontend (e.g. {"retry_after":5})
	cause      error          // wrap original
}

// ========================= factories =========================

func BadRequest(err error) *Error {
	return &Error{
		kind:       KindValidation,
		httpStatus: http.StatusBadRequest, // 400
		title:      "Malformed request.",
		detail:     "One or more fields are invalid.",
		code:       "validation_failed",
		cause:      err,
	}
}

func Unauthorized(err error) *Error {
	return &Error{
		kind:       KindUnauthorized,
		httpStatus: http.StatusUnauthorized, // 401
		title:      "Unauthorized.",
		detail:     "Access to this resource is unauthorized.",
		code:       "unauthorized",
		cause:      err,
	}
}

func Forbidden(err error) *Error {
	return &Error{
		kind:       KindForbidden,
		httpStatus: http.StatusForbidden, // 403
		title:      "Forbidden.",
		detail:     "Access to this resource is forbidden.",
		code:       "forbidden",
		cause:      err,
	}
}

func NotFound(err error) *Error {
	return &Error{
		kind:       KindNotFound,
		httpStatus: http.StatusNotFound, // 404
		title:      "Resource not found.",
		detail:     "The requested resource was not found.",
		code:       "not_found",
		cause:      err,
	}
}

func Conflict(err error) *Error {
	return &Error{
		kind:       KindConflict,
		httpStatus: http.StatusConflict, // 409
		title:      "Resource already in use.",
		detail:     "The requested resource already in use or is conflicting.",
		code:       "conflict",
		cause:      err,
	}
}

func TooManyRequests(err error) *Error {
	var (
		rateerr  *ratelimit.ErrRateLimited
		metadata map[string]any
	)

	if errors.As(err, &rateerr) {
		result := rateerr.Result()

		metadata = map[string]any{
			"limit":       result.Limit,
			"remaining":   result.Remaining,
			"reset":       result.ResetAfter.Seconds(),
			"retry_after": result.RetryAfter.Seconds(),
		}
	}

	return &Error{
		kind:       KindTooManyReq,
		httpStatus: http.StatusTooManyRequests, // 429
		title:      "Too many requests.",
		detail:     "Wait before request again.",
		code:       "too_many_requests",
		cause:      err,
		meta:       metadata,
	}
}

func Validation(fields []FieldError) *Error {
	return &Error{
		kind:       KindValidation,
		httpStatus: http.StatusUnprocessableEntity, // 422
		title:      "Invalid input.",
		detail:     "One or more fields are invalid.",
		fields:     fields,
		code:       "validation_failed",
	}
}

func Internal(err error) *Error {
	return &Error{
		kind:       KindInternal,
		httpStatus: http.StatusInternalServerError,
		title:      "Internal server error.",
		detail:     "Unexpected error.",
		code:       "internal_error",
		cause:      err,
	}
}

func BadGateway(err error) *Error {
	return &Error{
		kind:       KindBadGateway,
		httpStatus: http.StatusBadGateway,
		title:      "Bad gateway error.",
		detail:     "Bad gateway.",
		code:       "bad_gateway",
		cause:      err,
	}
}

// ========================= public getters =========================

func (e *Error) Error() string {
	return e.title + ": " + e.detail
}

func (e *Error) Unwrap() error { return e.cause }

func (e *Error) Title() string                { return e.title }
func (e *Error) Detail() string               { return e.detail }
func (e *Error) Meta() map[string]any         { return e.meta }
func (e *Error) Instance() string             { return e.instance }
func (e *Error) Kind() Kind                   { return e.kind }
func (e *Error) KindTitle() string            { return string(e.kind) }
func (e *Error) HTTPStatus() int              { return e.httpStatus }
func (e *Error) Code() string                 { return e.code }
func (e *Error) Fields() []FieldError         { return e.fields }
func (e *Error) WithInstance(i string) *Error { e.instance = i; return e }
func (e *Error) WithMeta(k string, v any) *Error {
	if e.meta == nil {
		e.meta = map[string]any{}
	}

	e.meta[k] = v

	return e
}

func IsKind(err error, k Kind) bool {
	var ae *Error
	if errors.As(err, &ae) {
		return ae.kind == k
	}

	return false
}

func From(err error) *Error {
	if err == nil {
		return nil
	}

	var ae *Error
	if errors.As(err, &ae) {
		return ae
	}

	return Internal(err)
}
