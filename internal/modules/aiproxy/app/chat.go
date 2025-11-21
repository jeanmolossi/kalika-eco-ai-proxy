package app

import (
	"context"
)

// Chat handles a full chat completion flow for a given tenant.
func (s *Service) Chat(ctx context.Context, in ChatInput) (ChatOutput, error) {
	req := in.Request
	req.TenantID = in.Tenant.ID
	req.Extras = in.Metadata

	// Pre guardrails.
	req, err := s.guardrails.PreProcessChat(ctx, in.Tenant, req)
	if err != nil {
		return ChatOutput{}, err
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

	// Post guardrails.
	resp, err = s.guardrails.PostProcessChat(ctx, in.Tenant, req, resp)
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
