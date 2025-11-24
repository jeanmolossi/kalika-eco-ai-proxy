package config

import "time"

// LLM groups configuration for the upstream LLM provider used as default.
// Tenants can override these values through their routing policy, but these
// settings ensure the proxy has a working provider even without per-tenant
// overrides.
type LLM struct {
	ProviderName    string        `env:"PROVIDER_NAME"    envDefault:"default-provider"`
	BaseURL         string        `env:"BASE_URL"         envDefault:"http://localhost:8080/v1"`
	APIKey          string        `env:"API_KEY"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT"  envDefault:"20s"`
	MaxRetries      int           `env:"MAX_RETRIES"      envDefault:"2"`
	EnableStreaming bool          `env:"ENABLE_STREAMING" envDefault:"true"`
	ChatModels      []string      `env:"CHAT_MODELS"      envSeparator:","`
	EmbedModels     []string      `env:"EMBED_MODELS"     envSeparator:","`
}
