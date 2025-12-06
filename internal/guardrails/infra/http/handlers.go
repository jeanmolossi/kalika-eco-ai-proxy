package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	"github.com/labstack/echo/v4"
)

// Handlers wires guardrails HTTP endpoints to the domain engine.
type Handlers struct {
	Engine  guardrails.Engine
	Tenants pkgtenant.Store
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(engine guardrails.Engine, tenants pkgtenant.Store) *Handlers {
	return &Handlers{Engine: engine, Tenants: tenants}
}

// EvaluateInput performs guardrail evaluation for request payloads.
func (h *Handlers) EvaluateInput(ctx echo.Context) error {
	var gx guardrails.Context
	if err := ctx.Bind(&gx); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := h.ensureTenant(ctx.Request().Context(), &gx); err != nil {
		return err
	}

	decision, err := h.Engine.EvaluateInput(ctx.Request().Context(), gx)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, decision)
}

// EvaluateOutput performs guardrail evaluation for response payloads.
func (h *Handlers) EvaluateOutput(ctx echo.Context) error {
	var gx guardrails.Context
	if err := ctx.Bind(&gx); err != nil {
		return ctx.NoContent(http.StatusBadRequest)
	}

	if err := h.ensureTenant(ctx.Request().Context(), &gx); err != nil {
		return err
	}

	decision, err := h.Engine.EvaluateOutput(ctx.Request().Context(), gx)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, decision)
}

func (h *Handlers) ensureTenant(ctx context.Context, gx *guardrails.Context) error {
	if gx.TenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required")
	}

	tenant, err := h.Tenants.FindByID(ctx, gx.TenantID)
	if err != nil {
		switch {
		case errors.Is(err, pkgtenant.ErrNotFound), errors.Is(err, pkgtenant.ErrInactiveTenant):
			return echo.NewHTTPError(http.StatusNotFound)
		default:
			return err
		}
	}

	if gx.Tags == nil {
		gx.Tags = map[string]string{}
	}

	if tenant != nil && tenant.PlanCode != "" {
		gx.Tags["tenant_plan"] = tenant.PlanCode
	}

	return nil
}
