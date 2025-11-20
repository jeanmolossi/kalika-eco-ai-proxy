package usage

import "context"

type Event struct {
	TenantID         string  `json:"tenant_id"`
	UserID           string  `json:"user_id,omitempty"`
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	RequestID        string  `json:"request_id"`
}

type Publisher interface {
	Publish(ctx context.Context, ev Event) error
}
