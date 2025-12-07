package http

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, handlers *Handlers) {
	tenants := g.Group("/tenants")

	tenants.GET("/api-keys/:apiKey", handlers.GetByAPIKey)
	tenants.GET("/:tenantID", handlers.GetByTenantID)
}
