package http

import (
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/gateway/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	pkgtokenizer "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tokenizer"
)

type Handlers struct {
	Tenants  pkgtenant.Store
	Limiter  ratelimit.Limiter
	Tokenizr pkgtokenizer.TokenCounter

	ChatUseCase       app.ChatUseCase
	EmbeddingsUseCase app.EmbeddingsUseCase
}

func NewHandlers(
	tenants pkgtenant.Store,
	limiter ratelimit.Limiter,
	tokenizr pkgtokenizer.TokenCounter,
	chat app.ChatUseCase,
	embeddings app.EmbeddingsUseCase,
) *Handlers {
	return &Handlers{
		Tenants:           tenants,
		Limiter:           limiter,
		Tokenizr:          tokenizr,
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
