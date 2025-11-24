package tenant

// RoutingConfig captures per-tenant upstream routing definitions.
// Each provider entry can supply its own base URL, credentials and model list.
type RoutingConfig struct {
	Providers []ProviderDefinition `json:"providers,omitempty"`
}

// ProviderDefinition describes how to reach a specific upstream provider.
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
