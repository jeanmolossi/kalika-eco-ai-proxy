package platform

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/cache"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
	"github.com/labstack/echo/v4"
)

// ModuleName is the identifier of this module.
const ModuleName = "platform"

// module implements core.Module and is responsible for the AI proxy data-plane.
type module struct{}

// NewModule creates a new AI proxy module.
func NewModule() core.Module {
	return &module{}
}

// Name implements core.Module.
func (m *module) Name() string { return ModuleName }

// Weight implements core.Module.
func (m *module) Weight() int { return 9 }

// Provide implements core.Module.
func (m *module) Provide(ctx context.Context, c *core.Container) error {
	conf := c.Config()
	logger := c.Logger()

	conn := core.MustGet[*pg.DB](c, database.PgConn)

	rl, err := ratelimit.NewLimiter(conf.RateLimit)
	if err != nil {
		return err
	}

	c.Set(core.TenantStoreModule, tenant.NewPostgresStore(conn.Pool()))
	c.Set(core.RateLimiterModule, rl)
	c.Set(core.SemanticCacheModule, cache.NewNoopSemanticCache())
	c.Set(core.GuardrailsModule, guardrails.NewNoopGuardrails())
	c.Set(core.UsagePublisherModule, usage.NewLogPublisher(logger))
	c.Set(core.AuditPublisherModule, audit.NewLogPublisher(logger))

	llmClient := llm.NewStubClient()
	c.Set(core.RouterModule, router.NewSimpleRouter(llmClient))

	return nil
}

// Routes implements core.Module.
func (m *module) Routes(e *echo.Echo, c *core.Container) error {
	return nil
}

// Start implements core.Module.
func (m *module) Start(ctx context.Context, c *core.Container) (stop func(context.Context) error, err error) {
	return stop, err
}

// Migrations implements core.Module.
func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
