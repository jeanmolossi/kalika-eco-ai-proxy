package tenant

type TenantConfig struct {
	ID             string
	Name           string
	Plan           string
	MaxTokensMonth int64
	ModelsAllowed  []string
	DefaultModel   string
	// policies de guardrail, cache, etc
	EnableSemanticCache bool
	MaxRequestsMinute   int64
}

type Store interface {
	FindByAPIKey(apiKey string) (*TenantConfig, error)
	FindByID(id string) (*TenantConfig, error)
}
