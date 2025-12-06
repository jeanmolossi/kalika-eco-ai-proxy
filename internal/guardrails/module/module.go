package guardrailsmodule

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/infra"
	guardhttp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/infra/http"
	"github.com/labstack/echo/v4"
)

// ModuleName is the identifier of this module.
const ModuleName = "guardrails"

type module struct{}

// NewModule creates a new guardrails module.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string { return ModuleName }
func (m *module) Weight() int  { return 3 }

func (m *module) Routes(g *echo.Group, c *core.Container) error {
	log := c.Logger()
	log.Info("guardrails: registering routes")

	deps := MustDepsFromContainer(c)

	handlers := guardhttp.NewHandlers(deps.Engine, deps.TenantStore)
	guardhttp.RegisterRoutes(g, handlers)

	return nil
}

func (m *module) Provide(ctx context.Context, c *core.Container) error {
	log := c.Logger()
	cfg := c.Config()

	log.Info("guardrails: providing dependencies", slog.String("env", cfg.Environment.String()))

	deps, err := buildDependencies(c)
	if err != nil {
		return err
	}

	c.Set(DepsKey, deps)
	c.Set(core.GuardrailsModule, deps.Engine)

	return nil
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return infra.Migrations(ctx, m)
}

// MigrationDB implements core.MigrationDBProvider ensuring guardrail migrations run in the guardrail database.
func (m *module) MigrationDB(_ context.Context, c *core.Container) (*sql.DB, error) {
	conn := core.MustGet[*pg.DB](c, database.GuardrailConn)
	return conn.SQL(), nil
}

// MustDepsFromContainer retrieves guardrails dependencies or panics if missing.
func MustDepsFromContainer(c *core.Container) Deps {
	v := c.MustGet(DepsKey)

	deps, ok := v.(Deps)
	if !ok {
		panic("guardrails: invalid deps type stored in container")
	}

	return deps
}
