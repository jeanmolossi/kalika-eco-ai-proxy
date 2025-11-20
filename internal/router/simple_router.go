package router

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
)

// SimpleRouter is a basic router implementation that forwards all requests
// to a single LLM client and optionally applies a default model from the
// tenant configuration.
type SimpleRouter struct {
	client llm.Client
}

// NewSimpleRouter creates a new SimpleRouter using the given LLM client.
func NewSimpleRouter(client llm.Client) *SimpleRouter {
	return &SimpleRouter{client: client}
}

// RouteChat chooses the effective model and forwards the request to the LLM client.
func (r *SimpleRouter) RouteChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatResponse, error) {
	// If no model is defined, fallback to default tenant model or stub.
	if req.Model == "" {
		if t.DefaultModel != "" {
			req.Model = t.DefaultModel
		} else {
			req.Model = "stub-model"
		}
	}

	return r.client.Chat(ctx, req)
}

// RouteEmbed forwards embedding requests to the LLM client.
func (r *SimpleRouter) RouteEmbed(ctx context.Context, t tenant.TenantConfig, req llm.EmbedRequest) (llm.EmbedResponse, error) {
	if req.Model == "" {
		req.Model = "stub-embed-model"
	}

	return r.client.Embed(ctx, req)
}
