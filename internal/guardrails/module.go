package guardrails

import (
	"context"
	"database/sql"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/infra"
	"github.com/labstack/echo/v4"
)

const ModuleName = "guardrails"

type module struct{}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string { return ModuleName }
func (m *module) Weight() int  { return 3 }
func (m *module) Routes(g *echo.Group, c *core.Container) error {
	registerRoutes(g, c)
	return nil
}

func (m *module) Provide(_ context.Context, c *core.Container) error {
	c.Set(core.GuardrailsModule, ProvideGuardrails(c))
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
