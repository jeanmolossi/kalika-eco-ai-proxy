package cache

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
)

// NoopSemanticCache is a semantic cache implementation that never hits
// and never stores anything. It is used as a safe default for development
// and testing.
type NoopSemanticCache struct{}

// NewNoopSemanticCache creates a new NoopSemanticCache instance.
func NewNoopSemanticCache() *NoopSemanticCache {
	return &NoopSemanticCache{}
}

// LookupChat always returns (nil, false, nil), meaning "no cached value".
func (c *NoopSemanticCache) LookupChat(ctx context.Context, tenantID string, req llm.ChatRequest) (*llm.ChatResponse, bool, error) {
	return nil, false, nil
}

// StoreChat is a no-op and ignores all values.
func (c *NoopSemanticCache) StoreChat(ctx context.Context, tenantID string, req llm.ChatRequest, resp llm.ChatResponse) error {
	return nil
}
