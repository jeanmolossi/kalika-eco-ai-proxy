package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	observability "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
)

// Embeddings handles an embeddings request for a given tenant.
func (s *Service) Embeddings(ctx context.Context, in EmbeddingsInput) (EmbeddingsOutput, error) {
	req := in.Request

	requestID := httpx.RequestIDFromCtx(ctx)
	if requestID == "" {
		requestID = uuid.NewString()
	}

	if s.tokenizr == nil {
		return EmbeddingsOutput{}, errors.New("tokenizer unavailable")
	}

	promptTokens, err := s.tokenizr.CountEmbeddingTokens(req.Model, req.Input)
	if err == nil {
		if res, err := s.limiter.Allow(ctx, in.Tenant.ID, "embeddings", promptTokens); err != nil {
			return EmbeddingsOutput{}, err
		} else if !res.Allowed {
			return EmbeddingsOutput{}, ErrRateLimited
		}
	}

	resp, err := s.router.RouteEmbed(ctx, in.Tenant, req)
	if err != nil {
		return EmbeddingsOutput{}, err
	}

	_ = s.usagePub.Publish(ctx, observability.UsageEvent{
		TenantID:         in.Tenant.ID,
		UserID:           in.UserID,
		Model:            resp.Model,
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		CostUSD:          observability.CalculateUSD(resp.Model, promptTokens, 0),
		RequestID:        requestID,
	})

	// You can extend audit here if needed.

	return resp, nil
}
