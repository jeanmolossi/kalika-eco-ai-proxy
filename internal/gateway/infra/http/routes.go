package http

import "github.com/labstack/echo/v4"

// RegisterRoutes wires all HTTP routes for the AI proxy module.
func RegisterRoutes(g *echo.Group, handler *Handlers) {
	// Public data-plane routes
	v1 := g.Group("/v1")
	v1.POST("/chat/completions", handler.ChatCompletions)
	v1.POST("/embeddings", handler.Embeddings)
	// You can also register health checks, etc., here if desired.
}
