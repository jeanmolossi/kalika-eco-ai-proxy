package tenant

import platformtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/app"

// Re-export tenant-facing domain contracts so that dependants can import
// a stable package path while the implementation lives under internal/.
type (
	PolicyConfig       = platformtenant.PolicyConfig
	RoutingConfig      = platformtenant.RoutingConfig
	ProviderDefinition = platformtenant.ProviderDefinition
	TenantConfig       = platformtenant.TenantConfig
	Store              = platformtenant.Store
)

var (
	ErrNotFound       = platformtenant.ErrNotFound
	ErrInvalidAPIKey  = platformtenant.ErrInvalidAPIKey
	ErrInactiveTenant = platformtenant.ErrInactiveTenant
)
