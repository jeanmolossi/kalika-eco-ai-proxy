package usage

import (
	"context"
	"log/slog"
)

// LogPublisher is a usage publisher that logs all usage events.
// It is useful during development to inspect usage behavior without
// depending on external infrastructure.
type LogPublisher struct {
	log *slog.Logger
}

// NewLogPublisher creates a new LogPublisher instance.
func NewLogPublisher(log *slog.Logger) *LogPublisher {
	return &LogPublisher{log: log}
}

// Publish logs the usage event and returns nil.
func (p *LogPublisher) Publish(ctx context.Context, ev Event) error {
	if p.log == nil {
		return nil
	}

	p.log.Info("usage event",
		slog.String("tenant_id", ev.TenantID),
		slog.String("user_id", ev.UserID),
		slog.String("model", ev.Model),
		slog.Int("prompt_tokens", ev.PromptTokens),
		slog.Int("completion_tokens", ev.CompletionTokens),
		slog.Float64("cost_usd", ev.CostUSD),
		slog.String("request_id", ev.RequestID),
	)

	return nil
}
