package tokenizer

import pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"

// NoopTokenCounter is a TokenCounter implementation that always returns zero.
// It is useful as a fallback or for test environments where token cost does not matter.
type NoopTokenCounter struct{}

// NewNoopTokenCounter creates a new NoopTokenCounter instance.
func NewNoopTokenCounter() *NoopTokenCounter {
	return &NoopTokenCounter{}
}

// CountChatTokens always returns 0 without error.
func (n *NoopTokenCounter) CountChatTokens(model string, messages []pkgllm.ChatMessage) (int, error) {
	return 0, nil
}

// CountEmbeddingTokens always returns 0 without error.
func (n *NoopTokenCounter) CountEmbeddingTokens(model string, inputs []string) (int, error) {
	return 0, nil
}
