package router

import (
	"context"
	"fmt"
	"slices"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

// SimpleRouter is a basic router implementation that forwards all requests
// to a single LLM client and optionally applies a default model from the
// tenant configuration.
type SimpleRouter struct {
	clients ClientPool
}

// NewSimpleRouter creates a new SimpleRouter using the given LLM client.
func NewSimpleRouter(pool ClientPool) *SimpleRouter {
	return &SimpleRouter{clients: pool}
}

// RouteChat chooses the effective model and forwards the request to the LLM client.
func (r *SimpleRouter) RouteChat(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.ChatRequest) (pkgllm.ChatResponse, error) {
	resolved, err := ResolveChatModel(t, req.Model)
	if err != nil {
		return pkgllm.ChatResponse{}, err
	}

	req.Model = resolved

	client, err := r.clients.ClientFor(ctx, t, resolved)
	if err != nil {
		return pkgllm.ChatResponse{}, err
	}

	return client.Chat(ctx, req)
}

// RouteEmbed forwards embedding requests to the LLM client.
func (r *SimpleRouter) RouteEmbed(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.EmbedRequest) (pkgllm.EmbedResponse, error) {
	resolved, err := ResolveEmbedModel(t, req.Model)
	if err != nil {
		return pkgllm.EmbedResponse{}, err
	}

	req.Model = resolved

	client, err := r.clients.ClientFor(ctx, t, resolved)
	if err != nil {
		return pkgllm.EmbedResponse{}, err
	}

	return client.Embed(ctx, req)
}

// ClientPool abstracts how we choose an LLM client for a tenant/model pair.
type ClientPool interface {
	ClientFor(ctx context.Context, t pkgtenant.TenantConfig, model string) (pkgllm.Client, error)
}

// ResolveChatModel enforces tenant allowlists and picks the effective chat model.
func ResolveChatModel(t pkgtenant.TenantConfig, requested string) (string, error) {
	allowed := allowedModels(t)
	if len(allowed) == 0 {
		return "", fmt.Errorf("no models allowed for tenant %s", t.ID)
	}

	candidate := requested
	if candidate == "" {
		candidate = t.DefaultModel
	}

	if candidate == "" {
		return "", fmt.Errorf("missing model and tenant default for %s", t.ID)
	}

	if !slices.Contains(allowed, candidate) {
		return "", fmt.Errorf("model %s not allowed for tenant %s", candidate, t.ID)
	}

	return candidate, nil
}

// ResolveEmbedModel enforces tenant allowlists and picks the effective embedding model.
func ResolveEmbedModel(t pkgtenant.TenantConfig, requested string) (string, error) {
	allowed := allowedModels(t)
	if len(allowed) == 0 {
		return "", fmt.Errorf("no models allowed for tenant %s", t.ID)
	}

	candidate := requested
	if candidate == "" {
		candidate = t.DefaultModel
	}

	if candidate == "" {
		return "", fmt.Errorf("missing embedding model for %s", t.ID)
	}

	if !slices.Contains(allowed, candidate) {
		return "", fmt.Errorf("model %s not allowed for tenant %s", candidate, t.ID)
	}

	return candidate, nil
}

func allowedModels(t pkgtenant.TenantConfig) []string {
	if t.ParsedPolicyConfig != nil && len(t.ParsedPolicyConfig.ModelsAllowed) > 0 {
		return t.ParsedPolicyConfig.ModelsAllowed
	}

	if t.DefaultModel != "" {
		return []string{t.DefaultModel}
	}

	return nil
}
