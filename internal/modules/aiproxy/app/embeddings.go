package app

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
)

// Embeddings handles an embeddings request for a given tenant.
func (s *Service) Embeddings(ctx context.Context, in EmbeddingsInput) (EmbeddingsOutput, error) {
	// Rate limit per tenant.
	ok, err := s.limiter.Allow(ctx, in.Tenant.ID, "embeddings", 1)
	if err != nil {
		return EmbeddingsOutput{}, err
	}

	if !ok {
		return EmbeddingsOutput{}, ErrRateLimited
	}

	req := in.Request
	// No guardrails or cache for MVP embeddings; can be added later.

	resp, err := s.router.RouteEmbed(ctx, in.Tenant, req)
	if err != nil {
		return EmbeddingsOutput{}, err
	}

	// Usage and audit for embeddings (minimal).
	_ = s.usagePub.Publish(ctx, usage.Event{
		TenantID:         in.Tenant.ID,
		UserID:           in.UserID,
		Model:            resp.Model,
		PromptTokens:     0,
		CompletionTokens: 0,
		CostUSD:          0,
		RequestID:        "",
	})

	// You can extend audit here if needed.

	return resp, nil
}
