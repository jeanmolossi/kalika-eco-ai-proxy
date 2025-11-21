package llm

import "context"

type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float32       `json:"temperature,omitempty"`
	TopP        float32       `json:"top_p,omitempty"`
	// metadata internos
	TenantID string            `json:"-"`
	UserID   string            `json:"-"`
	Extras   map[string]string `json:"-"`
}

type ChatResponse struct {
	ID        string        `json:"id"`
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	PromptTok int           `json:"prompt_tokens"`
	CompTok   int           `json:"completion_tokens"`
}

type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Model      string      `json:"model"`
	Embeddings [][]float32 `json:"embeddings"`
}

type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
}

const RoleUser = "user"
