package observability

import (
	"context"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
)

// AuditEvent records the prompt/response pair for policy and compliance auditing.
type AuditEvent struct {
	TenantID  string               `json:"tenant_id"`
	UserID    string               `json:"user_id,omitempty"`
	RequestID string               `json:"request_id"`
	Model     string               `json:"model"`
	Prompt    []pkgllm.ChatMessage `json:"prompt"`
	Response  []pkgllm.ChatMessage `json:"response"`
	Metadata  map[string]string    `json:"metadata,omitempty"`
}

// AuditPublisher is the contract for emitting audit events to an observability backend.
type AuditPublisher interface {
	Publish(ctx context.Context, ev AuditEvent) error
}
