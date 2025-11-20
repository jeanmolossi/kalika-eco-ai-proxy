package audit

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
)

type Event struct {
	TenantID  string            `json:"tenant_id"`
	UserID    string            `json:"user_id,omitempty"`
	RequestID string            `json:"request_id"`
	Model     string            `json:"model"`
	Prompt    []llm.ChatMessage `json:"prompt"`
	Response  []llm.ChatMessage `json:"response"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type Publisher interface {
	Publish(ctx context.Context, ev Event) error
}
