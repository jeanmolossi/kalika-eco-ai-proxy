package tenant

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrNotFound       = errors.New("tenant not found")
	ErrInvalidAPIKey  = errors.New("invalid api key")
	ErrInactiveTenant = errors.New("inactive tenant")
)

type PolicyConfig struct {
	ModelsAllowed []string        `json:"models_allowed,omitempty"`
	Routing       *RoutingConfig  `json:"routing,omitempty"`
	Guardrails    json.RawMessage `json:"guardrails,omitempty"`
}

type TenantConfig struct {
	ID                string
	Code              string
	Name              string
	Status            string
	PlanCode          string
	MaxTokensMonth    int64
	MaxRequestsMinute int64

	DefaultModel        string
	EnableSemanticCache bool
	CacheTTLSecs        int32
	MaxPromptTokens     int32
	MaxCompletionTokens int32
	PolicyConfigRaw     json.RawMessage
	ParsedPolicyConfig  *PolicyConfig
}

type Store interface {
	FindByAPIKey(ctx context.Context, apiKey string) (*TenantConfig, error)
	FindByID(ctx context.Context, tenantID string) (*TenantConfig, error)
	RevokeExpired(ctx context.Context) (int64, error)
}
