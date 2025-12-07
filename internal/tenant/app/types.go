package app

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

type RoutingConfig struct {
	Providers []ProviderDefinition `json:"providers,omitempty"`
}

type ProviderDefinition struct {
	Name             string   `json:"name"`
	BaseURL          string   `json:"base_url"`
	APIKey           string   `json:"api_key"`
	RequestTimeoutMS int      `json:"request_timeout_ms"`
	MaxRetries       int      `json:"max_retries"`
	EnableStreaming  bool     `json:"enable_streaming"`
	ChatModels       []string `json:"chat_models"`
	EmbedModels      []string `json:"embed_models"`
}
