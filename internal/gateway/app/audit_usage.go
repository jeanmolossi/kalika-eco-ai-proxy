package app

import (
	"context"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	observability "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
)

// publishUsage sends a usage event based on the LLM response.
func (s *Service) publishUsage(ctx context.Context, requestID string, in ChatInput, resp pkgllm.ChatResponse) error {
	cost := observability.CalculateUSD(resp.Model, resp.PromptTok, resp.CompTok)

	return s.usagePub.Publish(ctx, observability.UsageEvent{
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
func (s *Service) publishAudit(
	ctx context.Context,
	requestID string,
	in ChatInput,
	req pkgllm.ChatRequest,
	resp pkgllm.ChatResponse,
) error {
	return s.auditPub.Publish(ctx, observability.AuditEvent{
		TenantID:  in.Tenant.ID,
		UserID:    in.UserID,
		RequestID: requestID,
		Model:     resp.Model,
		Prompt:    req.Messages,
		Response:  resp.Messages,
		Metadata:  in.Metadata,
	})
}
