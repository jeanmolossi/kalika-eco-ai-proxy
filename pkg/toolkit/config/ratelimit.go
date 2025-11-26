package config

import "time"

// BackendType represents the type of rate limit backend.
type BackendType string

const (
	BackendNoop  BackendType = "noop"
	BackendLocal BackendType = "local"
)

type RateLimit struct {
	// Enabled globally enables or disables rate limiting.
	Enabled bool `env:"ENABLED" envDefault:"false"`

	// Backend defines which backend implementation should be used.
	// Supported values: "noop", "local".
	Backend BackendType `env:"BACKEND" envDefault:"noop"`

	// DefaultTokensPerInterval is the default number of tokens allowed per interval
	// for a given (tenantID, key) pair when no specific policy is provided.
	DefaultTokensPerInterval int `env:"DEFAULT_TOKENS_PER_INTERVAL" envDefault:"60"`

	// Interval defines the duration of the default rate-limiting window.
	Interval time.Duration `env:"INTERVAL" envDefault:"1m"`

	// MaxTokens defines the maximum bucket capacity for a key.
	// If zero or negative, it defaults to DefaultTokensPerInterval.
	MaxTokens int `env:"MAX_TOKENS"`
}
