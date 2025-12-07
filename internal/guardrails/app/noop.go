package app

import (
	"context"
)

// NoopGuardrails is a guardrails implementation that does not enforce
// any checks or modifications. It simply passes through the input and output.
type NoopGuardrails struct{}

// NewNoopGuardrails creates a new NoopGuardrails instance.
func NewNoopGuardrails() Engine {
	return &NoopGuardrails{}
}

// EvaluateInput implements Engine.
func (g *NoopGuardrails) EvaluateInput(ctx context.Context, gx Context) (Decision, error) {
	return Decision{}, nil
}

// EvaluateOutput implements Engine.
func (g *NoopGuardrails) EvaluateOutput(ctx context.Context, gx Context) (Decision, error) {
	return Decision{}, nil
}
