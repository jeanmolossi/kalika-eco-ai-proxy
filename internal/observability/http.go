package observability

import (
	"net/http"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/usage"
	"github.com/labstack/echo/v4"
)

func registerRoutes(g *echo.Group, c *core.Container) error {
	usagePub := core.MustGet[usage.Publisher](c, core.UsagePublisherModule)
	auditPub := core.MustGet[audit.Publisher](c, core.AuditPublisherModule)

	g.POST("/observability/usage", func(ctx echo.Context) error {
		var event usage.Event
		if err := ctx.Bind(&event); err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		if err := usagePub.Publish(ctx.Request().Context(), event); err != nil {
			return err
		}

		return ctx.NoContent(http.StatusAccepted)
	})

	g.POST("/observability/audit", func(ctx echo.Context) error {
		var event audit.Event
		if err := ctx.Bind(&event); err != nil {
			return ctx.NoContent(http.StatusBadRequest)
		}

		if err := auditPub.Publish(ctx.Request().Context(), event); err != nil {
			return err
		}

		return ctx.NoContent(http.StatusAccepted)
	})

	return nil
}
