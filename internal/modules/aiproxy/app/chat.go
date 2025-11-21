package app

import (
	"context"
	"errors"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/apperr"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
)

// Chat handles a full chat completion flow for a given tenant.
func (s *Service) Chat(ctx context.Context, in ChatInput) (ChatOutput, error) {
	req := in.Request
	req.TenantID = in.Tenant.ID
	req.Extras = in.Metadata
	now := time.Now()

	gctx := guardrails.Context{
		TenantID:      in.Tenant.ID,
		APIKeyID:      in.APIKey,
		Endpoint:      "chat.completions",
		Model:         req.Model,
		RequestID:     httpx.RequestIDFromCtx(ctx),
		UserID:        "",
		OccurredAt:    now,
		InputMessages: flattenChatMessages(req.Messages),
		Tags: map[string]string{
			"source": "aiproxy.chat",
		},
	}

	// Pre guardrails.
	decision, err := s.guardrails.EvaluateInput(ctx, gctx)
	if err != nil {
		return ChatOutput{}, err
	}

	switch decision.Action {
	case guardrails.ActionBlock:
		return ChatOutput{}, apperr.BadRequest(errors.New(decision.Reason))
	case guardrails.ActionRewrite:
		req.Messages = rebuildChatMessages(req.Messages, decision.RewrittenInputMessages)
	default:
	}

	// Optional semantic cache.
	if in.Tenant.EnableSemanticCache {
		if cached, ok, _ := s.cache.LookupChat(ctx, in.Tenant.ID, req); ok {
			_ = s.publishUsage(ctx, in, *cached)
			_ = s.publishAudit(ctx, in, req, *cached)
			return *cached, nil
		}
	}

	// Route to the appropriate model.
	resp, err := s.router.RouteChat(ctx, in.Tenant, req)
	if err != nil {
		return ChatOutput{}, err
	}

	gctx.OutputMessages = flattenChatMessages(resp.Messages)

	// Post guardrails.
	decision, err = s.guardrails.EvaluateOutput(ctx, gctx)
	if err != nil {
		return ChatOutput{}, err
	}

	// Persist cache when enabled.
	if in.Tenant.EnableSemanticCache {
		_ = s.cache.StoreChat(ctx, in.Tenant.ID, req, resp)
	}

	_ = s.publishUsage(ctx, in, resp)
	_ = s.publishAudit(ctx, in, req, resp)

	return resp, nil
}
