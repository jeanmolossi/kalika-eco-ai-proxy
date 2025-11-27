package guardrails

import (
"net/http"

"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
"github.com/labstack/echo/v4"
)

func registerRoutes(g *echo.Group, c *core.Container) {
engine := core.MustGet[Engine](c, core.GuardrailsModule)

g.POST("/guardrails/evaluate/input", func(ctx echo.Context) error {
var gx Context
if err := ctx.Bind(&gx); err != nil {
return ctx.NoContent(http.StatusBadRequest)
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

decision, err := engine.EvaluateOutput(ctx.Request().Context(), gx)
if err != nil {
return err
}

return ctx.JSON(http.StatusOK, decision)
})
}

