package observability

import (
	"context"
	"log/slog"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/infra"
	observabilityhttp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability/infra/http"
	"github.com/labstack/echo/v4"
)

const ModuleName = "observability"

type module struct{}

func NewModule() core.Module { return &module{} }

func (m *module) Name() string { return ModuleName }
func (m *module) Weight() int  { return 7 }
func (m *module) Routes(g *echo.Group, c *core.Container) error {
	log := c.Logger()
	log.Info("observability: registering routes")

	deps := MustDepsFromContainer(c)
	handlers := observabilityhttp.NewHandlers(deps.UsagePublisher, deps.AuditPublisher)

	observabilityhttp.RegisterRoutes(g, handlers)

	return nil
}

func (m *module) Provide(_ context.Context, c *core.Container) error {
	log := c.Logger()
	cfg := c.Config()

	log.Info("observability: providing dependencies", slog.String("env", cfg.Environment.String()))

	deps, err := buildDependencies(c)
	if err != nil {
		return err
	}

	c.Set(DepsKey, deps)
	c.Set(core.UsagePublisherModule, deps.UsagePublisher)
	c.Set(core.AuditPublisherModule, deps.AuditPublisher)

	return nil
}

func (m *module) Start(_ context.Context, c *core.Container) (func(context.Context) error, error) {
	deps := MustDepsFromContainer(c)

	return func(ctx context.Context) error {
		log := c.Logger()

		for _, closer := range deps.Closers {
			if err := closer.Close(); err != nil {
				log.ErrorContext(ctx, "close observability resource", "err", err)
			}
		}

		return nil
	}, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return infra.Migrations(ctx, m)
}
