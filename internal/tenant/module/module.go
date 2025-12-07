package tenantmodule

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	tenanthttp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/adapters/http"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/infra"
	"github.com/labstack/echo/v4"
)

const ModuleName = "tenant"

type module struct{}

// NewModule creates the tenant module.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string { return ModuleName }
func (m *module) Weight() int  { return 2 }

func (m *module) Routes(g *echo.Group, c *core.Container) error {
	log := c.Logger()
	log.Info("tenant: registering routes")

	deps := MustDepsFromContainer(c)
	handlers := tenanthttp.NewHandlers(deps.Store)

	tenanthttp.RegisterRoutes(g, handlers)

	return nil
}

func (m *module) Provide(ctx context.Context, c *core.Container) error {
	log := c.Logger()
	cfg := c.Config()

	log.Info("tenant: providing dependencies", slog.String("env", cfg.Environment.String()))

	deps, err := buildDependencies(c)
	if err != nil {
		return err
	}

	c.Set(DepsKey, deps)
	c.Set(core.TenantStoreModule, deps.Store)

	return nil
}

// MigrationDB implements core.MigrationDBProvider ensuring tenant migrations run on the tenant database.
func (m *module) MigrationDB(_ context.Context, c *core.Container) (*sql.DB, error) {
	conn := core.MustGet[*pg.DB](c, database.TenantConn)
	return conn.SQL(), nil
}

func (m *module) Start(ctx context.Context, c *core.Container) (func(context.Context) error, error) {
	deps := MustDepsFromContainer(c)
	log := c.Logger()

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Hour)

		defer func() {
			ticker.Stop()
			close(done)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if revoked, err := deps.Store.RevokeExpired(ctx); err != nil {
					log.ErrorContext(ctx, "api key revocation", "err", err)
				} else if revoked > 0 {
					log.InfoContext(ctx, "revoked expired api keys", "count", revoked)
				}
			}
		}
	}()

	return func(ctx context.Context) error {
		cancel()

		select {
		case <-done:
		case <-ctx.Done():
		}

		return nil
	}, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return infra.Migrations(ctx, m)
}

// MustDepsFromContainer retrieves tenant dependencies or panics.
func MustDepsFromContainer(c *core.Container) Deps {
	v := c.MustGet(DepsKey)

	deps, ok := v.(Deps)
	if !ok {
		panic("tenant: invalid deps type stored in container")
	}

	return deps
}
