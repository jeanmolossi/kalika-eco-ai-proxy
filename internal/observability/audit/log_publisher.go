package audit

import (
	"context"
	"log/slog"
)

// LogPublisher is an audit publisher that writes audit events to the logger.
// It should only be used in development, since prompts and responses may
// contain sensitive information.
type LogPublisher struct {
	log *slog.Logger
}

// NewLogPublisher creates a new LogPublisher instance.
func NewLogPublisher(log *slog.Logger) *LogPublisher {
	return &LogPublisher{log: log}
}

// Publish logs the audit event and returns nil.
func (p *LogPublisher) Publish(ctx context.Context, ev Event) error {
	if p.log == nil {
		return nil
	}

	p.log.Info("audit event",
		slog.String("tenant_id", ev.TenantID),
		slog.String("user_id", ev.UserID),
		slog.String("request_id", ev.RequestID),
		slog.String("model", ev.Model),
		slog.Int("prompt_len", len(ev.Prompt)),
		slog.Int("response_len", len(ev.Response)),
	)

	return nil
}
