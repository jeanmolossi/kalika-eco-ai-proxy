package app

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
)

// publishUsage sends a usage event based on the LLM response.
func (s *Service) publishUsage(ctx context.Context, requestID string, in ChatInput, resp llm.ChatResponse) error {
	cost := usage.CalculateUSD(resp.Model, resp.PromptTok, resp.CompTok)

	return s.usagePub.Publish(ctx, usage.Event{
		TenantID:         in.Tenant.ID,
		UserID:           in.UserID,
		Model:            resp.Model,
		PromptTokens:     resp.PromptTok,
		CompletionTokens: resp.CompTok,
		CostUSD:          cost,
		RequestID:        requestID,
	})
}

// publishAudit sends an audit event for the given request/response pair.
func (s *Service) publishAudit(ctx context.Context, requestID string, in ChatInput, req llm.ChatRequest, resp llm.ChatResponse) error {
	return s.auditPub.Publish(ctx, audit.Event{
		TenantID:  in.Tenant.ID,
		UserID:    in.UserID,
		RequestID: requestID,
		Model:     resp.Model,
		Prompt:    req.Messages,
		Response:  resp.Messages,
		Metadata:  in.Metadata,
	})
}
