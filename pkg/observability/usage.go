package observability

import "context"

// UsageEvent captures the usage metrics emitted by the gateway.
type UsageEvent struct {
	TenantID         string  `json:"tenant_id"`
	UserID           string  `json:"user_id,omitempty"`
	Model            string  `json:"model"`
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
	RequestID        string  `json:"request_id"`
}

// UsagePublisher is the contract for emitting usage events to an observability/billing backend.
type UsagePublisher interface {
	Publish(ctx context.Context, ev UsageEvent) error
}
