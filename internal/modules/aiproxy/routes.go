package aiproxy

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/httpapi"
	"github.com/labstack/echo/v4"
)

// registerRoutes wires all HTTP routes for the AI proxy module.
func registerRoutes(e *echo.Echo, deps Deps) {
	// Chat completions endpoint.
	chatHandler := &httpapi.ChatHandler{
		Tenants:    deps.TenantStore,
		Router:     deps.Router,
		Cache:      deps.Cache,
		Guardrails: deps.Guardrails,
		UsagePub:   deps.UsagePub,
		AuditPub:   deps.AuditPub,
		Limiter:    deps.Limiter,
	}

	// Embeddings handler – can follow the same pattern as ChatHandler.
	embedHandler := &httpapi.EmbeddingsHandler{
		Tenants: deps.TenantStore,
		Router:  deps.Router,
		// Or a direct LLM client if you prefer.
	}

	// Public data-plane routes:
	v1 := e.Group("/api/v1")
	v1.POST("/chat/completions", echo.WrapHandler(chatHandler))
	v1.POST("/embeddings", echo.WrapHandler(embedHandler))

	// You can also register health checks, etc., here if desired.
}
