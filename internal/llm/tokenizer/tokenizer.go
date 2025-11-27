package tokenizer

import pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"

// TokenCounter defines the interface for counting tokens for chat and embeddings.
// Implementations can be provider-specific (OpenAI, Anthropic, etc.).
type TokenCounter interface {
	// CountChatTokens returns the number of tokens used by the given messages for the specified model.
	CountChatTokens(model string, msgs []pkgllm.ChatMessage) (int, error)

	// CountEmbeddingTokens returns the number of tokens used by the given inputs for the specified model.
	CountEmbeddingTokens(model string, inputs []string) (int, error)
}

type TokenCounterResolver interface {
	Resolve(model string) (TokenCounter, error)
}
