package llm

import internalllm "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"

// Public-facing LLM contracts so callers depend on pkg instead of internal paths.
type (
	ChatMessage   = internalllm.ChatMessage
	ChatRequest   = internalllm.ChatRequest
	ChatResponse  = internalllm.ChatResponse
	EmbedRequest  = internalllm.EmbedRequest
	EmbedResponse = internalllm.EmbedResponse
	Client        = internalllm.Client
)

const (
	RoleUser      = internalllm.RoleUser
	RoleAssistant = internalllm.RoleAssistant
)
