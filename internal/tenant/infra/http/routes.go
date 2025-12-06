package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers tenant HTTP routes.
func RegisterRoutes(g *echo.Group, handlers *Handlers) {
	g.GET("/tenants/:tenantID", handlers.GetTenantByID)
	g.GET("/tenants/api-keys/:apiKey", handlers.GetTenantByAPIKey)
}
