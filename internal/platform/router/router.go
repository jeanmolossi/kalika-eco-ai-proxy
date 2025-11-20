package router

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
)

type Router interface {
	RouteChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatResponse, error)
	RouteEmbed(ctx context.Context, t tenant.TenantConfig, req llm.EmbedRequest) (llm.EmbedResponse, error)
}
