package app

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
)

// publishUsage sends a usage event based on the LLM response.
func (s *Service) publishUsage(ctx context.Context, in ChatInput, resp llm.ChatResponse) error {
	return s.usagePub.Publish(ctx, usage.Event{
		TenantID:         in.Tenant.ID,
		UserID:           in.UserID,
		Model:            resp.Model,
		PromptTokens:     resp.PromptTok,
		CompletionTokens: resp.CompTok,
		CostUSD:          0, // for now, stub cost
		RequestID:        "",
	})
}

// publishAudit sends an audit event for the given request/response pair.
func (s *Service) publishAudit(ctx context.Context, in ChatInput, req llm.ChatRequest, resp llm.ChatResponse) error {
	return s.auditPub.Publish(ctx, audit.Event{
		TenantID:  in.Tenant.ID,
		UserID:    in.UserID,
		RequestID: "",
		Model:     resp.Model,
		Prompt:    req.Messages,
		Response:  resp.Messages,
		Metadata:  in.Metadata,
	})
}
