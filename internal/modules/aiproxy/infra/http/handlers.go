package http

import (
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
)

type Handlers struct {
	Tenants tenant.Store

	ChatUseCase       app.ChatUseCase
	EmbeddingsUseCase app.EmbeddingsUseCase
}

func NewHandlers(
	tenants tenant.Store,
	chat app.ChatUseCase,
	embeddings app.EmbeddingsUseCase,
) *Handlers {
	return &Handlers{
		Tenants:           tenants,
		ChatUseCase:       chat,
		EmbeddingsUseCase: embeddings,
	}
}

// extractAPIKey parses the Authorization header and returns the API key.
// It supports the "Bearer <token>" format and also accepts a raw value.
func extractAPIKey(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	// Bearer token format
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	// Fallback: treat the header as the raw key.
	return authHeader
}
