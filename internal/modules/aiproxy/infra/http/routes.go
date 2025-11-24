package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes wires all HTTP routes for the AI proxy module.
func RegisterRoutes(e *echo.Echo, basePath string, handler *Handlers) {
	// Public data-plane routes (explicit base path to match the HTTP server prefix).
	v1 := e.Group(basePath)
	v1.POST("/chat/completions", handler.ChatCompletions)
	v1.POST("/embeddings", handler.Embeddings)
	// You can also register health checks, etc., here if desired.
}
