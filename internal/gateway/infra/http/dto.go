package http

import pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"

type chatRequestDTO struct {
	Model       string               `json:"model"`
	Messages    []pkgllm.ChatMessage `json:"messages"`
	MaxTokens   int                  `json:"max_tokens,omitempty"`
	Temperature float32              `json:"temperature,omitempty"`
	TopP        float32              `json:"top_p,omitempty"`
	Metadata    map[string]string    `json:"metadata,omitempty"`
}

type embedRequestDTO struct {
	Model    string            `json:"model"`
	Input    []string          `json:"input"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
