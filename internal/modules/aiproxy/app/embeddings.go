package app

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
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

	_ = s.usagePub.Publish(ctx, usage.Event{
		TenantID:         in.Tenant.ID,
		UserID:           in.UserID,
		Model:            resp.Model,
		PromptTokens:     promptTokens,
		CompletionTokens: 0,
		CostUSD:          usage.CalculateUSD(resp.Model, promptTokens, 0),
		RequestID:        requestID,
	})

	// You can extend audit here if needed.

	return resp, nil
}
