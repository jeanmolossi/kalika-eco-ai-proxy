package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers guardrails HTTP routes.
func RegisterRoutes(g *echo.Group, handlers *Handlers) {
	g.POST("/guardrails/evaluate/input", handlers.EvaluateInput)
	g.POST("/guardrails/evaluate/output", handlers.EvaluateOutput)
}
