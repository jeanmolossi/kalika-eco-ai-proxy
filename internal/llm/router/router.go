package router

import (
	"context"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

type Router interface {
	RouteChat(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.ChatRequest) (pkgllm.ChatResponse, error)
	RouteEmbed(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.EmbedRequest) (pkgllm.EmbedResponse, error)
}
