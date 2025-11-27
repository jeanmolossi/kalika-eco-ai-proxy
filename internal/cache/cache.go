package cache

import (
	"context"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
)

type SemanticCache interface {
	LookupChat(ctx context.Context, tenantID string, req pkgllm.ChatRequest) (*pkgllm.ChatResponse, bool, error)
	StoreChat(ctx context.Context, tenantID string, req pkgllm.ChatRequest, resp pkgllm.ChatResponse) error
}
