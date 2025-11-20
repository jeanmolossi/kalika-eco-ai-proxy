package guardrails

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
)

type Guardrails interface {
	PreProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatRequest, error)
	PostProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest, resp llm.ChatResponse) (llm.ChatResponse, error)
}
