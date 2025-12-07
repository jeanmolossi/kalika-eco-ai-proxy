package http

import (
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/app/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/app/usage"
	"github.com/labstack/echo/v4"
)

// Handlers exposes observability endpoints.
type Handlers struct {
	UsagePublisher usage.Publisher
	AuditPublisher audit.Publisher
}

// NewHandlers builds a new Handlers instance.
func NewHandlers(usagePub usage.Publisher, auditPub audit.Publisher) *Handlers {
	return &Handlers{UsagePublisher: usagePub, AuditPublisher: auditPub}
}

// PublishUsage receives and forwards usage events.
func (h *Handlers) PublishUsage(ctx echo.Context) error {
	var event usage.Event
	if err := ctx.Bind(&event); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := h.UsagePublisher.Publish(ctx.Request().Context(), event); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusAccepted)
}

// PublishAudit receives and forwards audit events.
func (h *Handlers) PublishAudit(ctx echo.Context) error {
	var event audit.Event
	if err := ctx.Bind(&event); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := h.AuditPublisher.Publish(ctx.Request().Context(), event); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusAccepted)
}
