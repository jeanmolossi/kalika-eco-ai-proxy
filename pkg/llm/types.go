package llm

import "context"

// Public-facing LLM contracts so callers depend on pkg instead of internal paths.
type (
	ChatMessage struct {
		Role    string `json:"role"` // "system", "user", "assistant"
		Content string `json:"content"`
	}

	ChatRequest struct {
		Model       string        `json:"model"`
		Messages    []ChatMessage `json:"messages"`
		MaxTokens   int           `json:"max_tokens,omitempty"`
		Temperature float32       `json:"temperature,omitempty"`
		TopP        float32       `json:"top_p,omitempty"`
		Stream      bool          `json:"stream,omitempty"`
		// metadata internos
		TenantID string            `json:"-"`
		UserID   string            `json:"-"`
		Extras   map[string]string `json:"-"`
	}

	ChatResponse struct {
		ID        string        `json:"id"`
		Model     string        `json:"model"`
		Messages  []ChatMessage `json:"messages"`
		PromptTok int           `json:"prompt_tokens"`
		CompTok   int           `json:"completion_tokens"`
	}

	EmbedRequest struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}

	EmbedResponse struct {
		Model      string      `json:"model"`
		Embeddings [][]float32 `json:"embeddings"`
	}

	Client interface {
		Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
		Embed(ctx context.Context, req EmbedRequest) (EmbedResponse, error)
	}
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
)
