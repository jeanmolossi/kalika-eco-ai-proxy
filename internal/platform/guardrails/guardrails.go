package guardrails

import (
	"context"
	"time"
)

type Phase string

const (
	PhaseInput  Phase = "input"
	PhaseOutput Phase = "output"
)

type Action string

const (
	ActionAllow   Action = "allow"   // segue o fluxo normal
	ActionBlock   Action = "block"   // retorna erro pro cliente
	ActionRewrite Action = "rewrite" // usa conteúdo reescrito
)

// Contexto genérico do proxy pra guardrails
type Context struct {
	TenantID   string
	APIKeyID   string
	Endpoint   string // ex: "chat.completions", "embeddings"
	Model      string
	UserID     string // se você tiver isso
	RequestID  string
	OccurredAt time.Time
	// Payload em formato neutro pra não acoplar de cara ao llm.ChatMessage
	// Você monta isso no app (chat/embeddings).
	InputMessages  []string // ex: todas as mensagens concatenadas ou normalizadas
	OutputMessages []string // preenchido só na fase de output
	// Opcional: tags extras
	Tags map[string]string
}

// Resultado da avaliação dos guardrails
type Decision struct {
	Action  Action
	Reason  string   // motivo aplicado (pra log/audit)
	RuleIDs []string // IDs das regras que bateram (pra auditoria futura)
	// Se Action == Rewrite, esses campos devem ser usados
	RewrittenInputMessages  []string
	RewrittenOutputMessages []string
	// Telemetria adicional, se quiser logar
	Metadata map[string]any
}

// Engine é o que o app vai usar
type Engine interface {
	EvaluateInput(ctx context.Context, gx Context) (Decision, error)
	EvaluateOutput(ctx context.Context, gx Context) (Decision, error)
}

type RuleKind string

const (
	RuleKindRegexBlock   RuleKind = "regex_block"
	RuleKindRegexRewrite RuleKind = "regex_rewrite"
	RuleKindMaxLength    RuleKind = "max_length"
)

type RuleConfig struct {
	Phase       Phase    `json:"phase"`
	Action      Action   `json:"action"`
	Pattern     string   `json:"pattern,omitempty"` // regex ou valor numérico em string
	Replacement string   `json:"Replacement,omitempty"`
	MaxLength   *int     `json:"max_length,omitempty"`
	Severity    string   `json:"severity,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type Rule struct {
	ID       string
	TenantID string
	Name     string
	Kind     RuleKind
	IsActive bool
	Priority int

	Config RuleConfig
}

type RuleRepository interface {
	ListRulesForTenantPhase(ctx context.Context, tenantID string, phase Phase) ([]Rule, error)
}
