package guardrails

import (
	"context"
	"log/slog"
)

type LoggerSink struct {
	logger *slog.Logger
}

func NewLoggerSink(logger *slog.Logger) *LoggerSink {
	return &LoggerSink{
		logger: logger.With("component", "guardrails.decision"),
	}
}

func (s *LoggerSink) RecordDecision(ctx context.Context, gx Context, phase Phase, dec Decision) {
	ev := buildDecisionEvent(gx, phase, dec)

	s.logger.InfoContext(ctx, "guardrail_decision",
		slog.String("tenant_id", ev.TenantID),
		slog.String("api_key_id", ev.APIKeyID),
		slog.String("user_id", ev.UserID),
		slog.String("request_id", ev.RequestID),
		slog.String("endpoint", ev.Endpoint),
		slog.String("model", ev.Model),
		slog.String("phase", string(ev.Phase)),
		slog.String("action", string(ev.Action)),
		slog.String("reason", ev.Reason),
		slog.String("severity", ev.Severity),
		slog.Any("rule_ids", ev.RuleIDs),
		slog.Any("tags", ev.Tags),
		slog.Int("input_size_bytes", ev.InputSizeBytes),
		slog.Int("output_size_bytes", ev.OutputSizeBytes),
		slog.Int("input_msg_count", ev.InputMsgCount),
		slog.Int("output_msg_count", ev.OutputMsgCount),
		slog.String("tenant_plan", ev.TenantPlan),
		slog.String("environment", ev.Environment),
		slog.String("direction", ev.Direction),
	)
}

func buildDecisionEvent(gx Context, phase Phase, dec Decision) DecisionEvent {
	inputBytes := 0
	for _, m := range gx.InputMessages {
		inputBytes += len(m)
	}

	outputBytes := 0
	for _, m := range gx.OutputMessages {
		outputBytes += len(m)
	}

	// juntar tags de todas as regras, se você tiver isso no Decision.Metadata depois
	var tags []string

	if raw, ok := dec.Metadata["tags"]; ok {
		if t, ok := raw.([]string); ok {
			tags = t
		}
	}

	ev := DecisionEvent{
		TenantID:   gx.TenantID,
		APIKeyID:   gx.APIKeyID,
		UserID:     gx.UserID,
		RequestID:  gx.RequestID,
		Endpoint:   gx.Endpoint,
		Model:      gx.Model,
		OccurredAt: gx.OccurredAt,

		Phase:    phase,
		Action:   dec.Action,
		Reason:   dec.Reason,
		RuleIDs:  dec.RuleIDs,
		Severity: "", // se você quiser derivar da regra dominante depois
		Tags:     tags,

		InputSizeBytes:  inputBytes,
		OutputSizeBytes: outputBytes,
		InputMsgCount:   len(gx.InputMessages),
		OutputMsgCount:  len(gx.OutputMessages),

		TenantPlan:  gx.Tags["tenant_plan"],
		Environment: gx.Tags["environment"],
	}

	if phase == PhaseInput {
		ev.Direction = "request"
	} else {
		ev.Direction = "response"
	}

	return ev
}
