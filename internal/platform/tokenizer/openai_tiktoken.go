package tokenizer

import (
	"fmt"

	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	"github.com/pkoukk/tiktoken-go"
)

// OpenAITikTokenCounter is a TokenCounter implementation based on OpenAI's tiktoken encoding.
// It should be used for models that are compatible with OpenAI tokenizers.
type OpenAITikTokenCounter struct {
	// modelAlias allows mapping custom model names to known OpenAI models for encoding purposes.
	modelAlias map[string]string
}

// NewOpenAITikTokenCounter creates a new OpenAITikTokenCounter instance.
// The alias map can be used to map internal model names to OpenAI model identifiers.
// For example: map["stub-model"] = "gpt-4o-mini"
func NewOpenAITikTokenCounter(alias map[string]string) *OpenAITikTokenCounter {
	if alias == nil {
		alias = make(map[string]string)
	}

	return &OpenAITikTokenCounter{
		modelAlias: alias,
	}
}

// resolveModel maps a given model to a known tokenizer model, if needed.
func (c *OpenAITikTokenCounter) resolveModel(model string) string {
	if mapped, ok := c.modelAlias[model]; ok {
		return mapped
	}

	return model
}

// CountChatTokens counts the tokens for a chat completion request using tiktoken.
func (c *OpenAITikTokenCounter) CountChatTokens(model string, messages []pkgllm.ChatMessage) (int, error) {
	model = c.resolveModel(model)

	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, fmt.Errorf("openai tokenizer: failed to load encoding for model %s: %w", model, err)
	}

	total := 0

	for _, msg := range messages {
		// Count role tokens
		total += len(enc.Encode(msg.Role, nil, nil))
		// Count content tokens
		total += len(enc.Encode(msg.Content, nil, nil))
	}

	return total, nil
}

// CountEmbeddingTokens counts the tokens for embeddings requests using tiktoken.
func (c *OpenAITikTokenCounter) CountEmbeddingTokens(model string, inputs []string) (int, error) {
	model = c.resolveModel(model)

	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return 0, fmt.Errorf("openai tokenizer: failed to load encoding for model %s: %w", model, err)
	}

	total := 0

	for _, in := range inputs {
		total += len(enc.Encode(in, nil, nil))
	}

	return total, nil
}
