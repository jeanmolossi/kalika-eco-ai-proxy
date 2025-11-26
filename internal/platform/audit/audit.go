package audit

import (
	"context"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
)

type Event struct {
	TenantID  string               `json:"tenant_id"`
	UserID    string               `json:"user_id,omitempty"`
	RequestID string               `json:"request_id"`
	Model     string               `json:"model"`
	Prompt    []pkgllm.ChatMessage `json:"prompt"`
	Response  []pkgllm.ChatMessage `json:"response"`
	Metadata  map[string]string    `json:"metadata,omitempty"`
}

type Publisher interface {
	Publish(ctx context.Context, ev Event) error
}
