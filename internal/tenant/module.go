package tenant

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/infra"
	"github.com/labstack/echo/v4"
)

const ModuleName = "tenant"

// module wires tenant storage and background maintenance tasks.
type module struct{}

// NewModule creates the tenant module.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string { return ModuleName }
func (m *module) Weight() int  { return 2 }
func (m *module) Routes(g *echo.Group, c *core.Container) error {
	store := core.MustGet[Store](c, core.TenantStoreModule)

	g.GET("/tenants/:tenantID", func(ctx echo.Context) error {
		tenantID := ctx.Param("tenantID")

		tenant, err := store.FindByID(ctx.Request().Context(), tenantID)
		if err != nil {
			if err == ErrNotFound {
				return ctx.NoContent(http.StatusNotFound)
			}

			return err
		}

		return ctx.JSON(http.StatusOK, tenant)
	})

	g.GET("/tenants/api-keys/:apiKey", func(ctx echo.Context) error {
		apiKey := ctx.Param("apiKey")

		tenant, err := store.FindByAPIKey(ctx.Request().Context(), apiKey)
		if err != nil {
			switch err {
			case ErrInvalidAPIKey, ErrNotFound, ErrInactiveTenant:
				return ctx.NoContent(http.StatusNotFound)
			default:
				return err
			}
		}

		return ctx.JSON(http.StatusOK, tenant)
	})

	return nil
}

func (m *module) Provide(_ context.Context, c *core.Container) error {
	conn := core.MustGet[*pg.DB](c, database.TenantConn)
	c.Set(core.TenantStoreModule, NewPostgresStore(conn.Pool()))
	return nil
}

// MigrationDB implements core.MigrationDBProvider ensuring tenant migrations run on the tenant database.
func (m *module) MigrationDB(_ context.Context, c *core.Container) (*sql.DB, error) {
	conn := core.MustGet[*pg.DB](c, database.TenantConn)
	return conn.SQL(), nil
}

func (m *module) Start(ctx context.Context, c *core.Container) (func(context.Context) error, error) {
	store := core.MustGet[Store](c, core.TenantStoreModule)
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
				if revoked, err := store.RevokeExpired(ctx); err != nil {
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
