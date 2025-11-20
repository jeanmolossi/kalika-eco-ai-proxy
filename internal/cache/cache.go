package cache

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
)

type SemanticCache interface {
	LookupChat(ctx context.Context, tenantID string, req llm.ChatRequest) (*llm.ChatResponse, bool, error)
	StoreChat(ctx context.Context, tenantID string, req llm.ChatRequest, resp llm.ChatResponse) error
}
