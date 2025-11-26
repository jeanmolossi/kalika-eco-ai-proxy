package aiproxy

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy/app"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	pkgtokenizer "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tokenizer"
)

// DepsKey is the container key used to store AI proxy dependencies.
const DepsKey = "aiproxy:deps"

// Deps groups all dependencies required by the AI proxy HTTP layer.
type Deps struct {
	TenantStore pkgtenant.Store
	Limiter     ratelimit.Limiter
	Tokenizr    pkgtokenizer.TokenCounter
	Service     *app.Service
}

// MustDepsFromContainer retrieves Deps from container or panics.
func MustDepsFromContainer(c *core.Container) Deps {
	v := c.MustGet(DepsKey)

	deps, ok := v.(Deps)
	if !ok {
		panic("aiproxy: invalid deps type stored in container")
	}

	return deps
}

// buildDependencies constructs the Deps struct with minimal stub implementations.
// This is meant to be the MVP wiring and can be replaced with real implementations over time.
func buildDependencies(c *core.Container) (Deps, error) {
	// For the MVP, create very simple noop/stub implementations.
	// You can later replace these with real DB/Redis/Kafka/etc. integrations.
	tenantStore := core.MustGet[pkgtenant.Store](c, core.TenantStoreModule)
	limiter := core.MustGet[app.TokenLimiter](c, core.RateLimiterModule)
	semCache := core.MustGet[app.SemanticCache](c, core.SemanticCacheModule)
	guard := core.MustGet[app.ChatGuardrails](c, core.GuardrailsModule)
	usagePub := core.MustGet[app.UsagePublisher](c, core.UsagePublisherModule)
	auditPub := core.MustGet[app.AuditPublisher](c, core.AuditPublisherModule)
	router := core.MustGet[app.ChatRouter](c, core.RouterModule)
	tokenizr := core.MustGet[app.TokenCounter](c, core.TokenizerModule)

	svc := app.NewService(
		limiter,
		semCache,
		guard,
		router,
		usagePub,
		auditPub,
		tokenizr,
	)

	deps := Deps{
		TenantStore: tenantStore,
		Limiter:     limiter,
		Tokenizr:    tokenizr,
		Service:     svc,
	}

	return deps, nil
}
