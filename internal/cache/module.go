package cache

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/labstack/echo/v4"
)

const ModuleName = "cache"

type module struct{}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 5 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	c.Set(core.SemanticCacheModule, NewNoopSemanticCache())
	return nil
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
