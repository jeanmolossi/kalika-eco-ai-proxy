package tenantmodule

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	tenantapp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/app"
	tenantinfra "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/infra"
)

// DepsKey is the container key used to store tenant dependencies.
const DepsKey = "tenant:deps"

// Deps groups dependencies offered by the tenant module.
type Deps struct {
	Store tenantapp.Store
}

func buildDependencies(c *core.Container) (Deps, error) {
	conn := core.MustGet[*pg.DB](c, database.TenantConn)

	deps := Deps{
		Store: tenantinfra.NewPostgresStore(conn.Pool()),
	}

	return deps, nil
}
