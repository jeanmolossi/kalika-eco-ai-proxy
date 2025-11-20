package http

import "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"

type chatRequestDTO struct {
	Model       string            `json:"model"`
	Messages    []llm.ChatMessage `json:"messages"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	Temperature float32           `json:"temperature,omitempty"`
	TopP        float32           `json:"top_p,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type embedRequestDTO struct {
	Model    string            `json:"model"`
	Input    []string          `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type embedVectorDTO struct {
	Index     int       `json:"index"`
	Embedding []float32 `json:"embedding"`
}

type embedResponseDTO struct {
	Model string           `json:"model"`
	Data  []embedVectorDTO `json:"data"`
}
