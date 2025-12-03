package llm

import pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"

type (
	ChatMessage   = pkgllm.ChatMessage
	ChatRequest   = pkgllm.ChatRequest
	ChatResponse  = pkgllm.ChatResponse
	EmbedRequest  = pkgllm.EmbedRequest
	EmbedResponse = pkgllm.EmbedResponse
	Client        = pkgllm.Client
)

const (
	RoleUser      = pkgllm.RoleUser
	RoleAssistant = pkgllm.RoleAssistant
)
