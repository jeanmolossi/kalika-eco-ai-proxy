package ratelimit

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	platformratelimit "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	"github.com/labstack/echo/v4"
)

const ModuleName = "ratelimit"

type module struct{}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 4 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	rl, err := platformratelimit.NewLimiter(c.Config().RateLimit)
	if err != nil {
		return err
	}

	c.Set(core.RateLimiterModule, rl)

	return nil
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
