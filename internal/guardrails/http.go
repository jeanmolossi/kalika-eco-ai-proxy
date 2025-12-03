package guardrails

import (
	"context"
	"errors"
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	"github.com/labstack/echo/v4"
)

func registerRoutes(g *echo.Group, c *core.Container) {
	engine := core.MustGet[Engine](c, core.GuardrailsModule)
	tenantStore := core.MustGet[pkgtenant.Store](c, core.TenantStoreModule)

	g.POST("/guardrails/evaluate/input", func(ctx echo.Context) error {
		var gx Context
		if err := ctx.Bind(&gx); err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		if err := ensureTenant(ctx.Request().Context(), tenantStore, &gx); err != nil {
			return err
		}

		decision, err := engine.EvaluateInput(ctx.Request().Context(), gx)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, decision)
	})

	g.POST("/guardrails/evaluate/output", func(ctx echo.Context) error {
		var gx Context
		if err := ctx.Bind(&gx); err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		if err := ensureTenant(ctx.Request().Context(), tenantStore, &gx); err != nil {
			return err
		}

		decision, err := engine.EvaluateOutput(ctx.Request().Context(), gx)
		if err != nil {
			return err
		}

		return ctx.JSON(http.StatusOK, decision)
	})
}

func ensureTenant(ctx context.Context, tenants pkgtenant.Store, gx *Context) error {
	if gx.TenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required")
	}

	tenant, err := tenants.FindByID(ctx, gx.TenantID)
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
