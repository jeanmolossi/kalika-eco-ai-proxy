package guardrails

import internal "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"

// Public contract for guardrails domain used by other bounded contexts.
type (
	Phase    = internal.Phase
	Action   = internal.Action
	Context  = internal.Context
	Decision = internal.Decision

	DecisionEvent = internal.DecisionEvent

	Engine         = internal.Engine
	RuleKind       = internal.RuleKind
	RuleConfig     = internal.RuleConfig
	Rule           = internal.Rule
	RuleRepository = internal.RuleRepository
	DecisionSink   = internal.DecisionSink
)

const (
	PhaseInput  Phase = internal.PhaseInput
	PhaseOutput Phase = internal.PhaseOutput

	ActionAllow   Action = internal.ActionAllow
	ActionBlock   Action = internal.ActionBlock
	ActionRewrite Action = internal.ActionRewrite

	RuleKindRegexBlock   RuleKind = internal.RuleKindRegexBlock
	RuleKindRegexRewrite RuleKind = internal.RuleKindRegexRewrite
	RuleKindMaxLength    RuleKind = internal.RuleKindMaxLength
)
