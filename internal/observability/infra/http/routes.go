package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers observability HTTP routes.
func RegisterRoutes(g *echo.Group, handlers *Handlers) {
	g.POST("/observability/usage", handlers.PublishUsage)
	g.POST("/observability/audit", handlers.PublishAudit)
}
