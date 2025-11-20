package database

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database/pg"
	"github.com/labstack/echo/v4"
)

// ModuleName is the identifier of this module.
const (
	ModuleName = "database"
	PgConn     = ModuleName + ":pgconn"
)

// module implements core.Module and is responsible for the AI proxy data-plane.
type module struct{}

// NewModule creates a new AI proxy module.
func NewModule() core.Module {
	return &module{}
}

// Name returns the module name for logging and debugging.
func (m *module) Name() string { return ModuleName }

// Weight controls the initialization order; lower values are started earlier.
// AI proxy depends on core infra (config, logging, etc.), so it can be a low, but not the lowest, value.
func (m *module) Weight() int { return 1 }

// Provide implements core.Module.
func (m *module) Provide(ctx context.Context, c *core.Container) error {
	cfg := c.Config()

	pgdb, err := pg.Open(ctx, cfg.PgDB)
	if err != nil {
		return err
	}

	c.Set(PgConn, pgdb)

	return nil
}

// Routes implements core.Module.
func (m *module) Routes(e *echo.Echo, c *core.Container) error { return nil }

// Start implements core.Module.
func (m *module) Start(ctx context.Context, c *core.Container) (stop func(context.Context) error, err error) {
	return stop, err
}

// Migrations implements core.Module.
func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
