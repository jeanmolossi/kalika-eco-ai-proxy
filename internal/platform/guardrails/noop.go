package guardrails

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
)

// NoopGuardrails is a guardrails implementation that does not enforce
// any checks or modifications. It simply passes through the input and output.
type NoopGuardrails struct{}

// NewNoopGuardrails creates a new NoopGuardrails instance.
func NewNoopGuardrails() *NoopGuardrails {
	return &NoopGuardrails{}
}

// PreProcessChat returns the original request without changes.
func (g *NoopGuardrails) PreProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatRequest, error) {
	return req, nil
}

// PostProcessChat returns the original response without changes.
func (g *NoopGuardrails) PostProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest, resp llm.ChatResponse) (llm.ChatResponse, error) {
	return resp, nil
}
