package guardrailsmodule

import (
	"fmt"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	guardrailsadapters "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/adapters"
	guardrailsapp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/app"
	guardrailsinfra "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails/infra"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

// DepsKey is the container key used to store guardrails dependencies.
const DepsKey = "guardrails:deps"

// Deps groups all dependencies required by the guardrails module.
type Deps struct {
	Engine      guardrailsapp.Engine
	TenantStore pkgtenant.Store
}

func buildDependencies(c *core.Container) (Deps, error) {
	engine, err := buildEngine(c)
	if err != nil {
		return Deps{}, err
	}

	tenantStore := core.MustGet[pkgtenant.Store](c, core.TenantStoreModule)

	deps := Deps{
		Engine:      engine,
		TenantStore: tenantStore,
	}

	return deps, nil
}

func buildEngine(c *core.Container) (guardrailsapp.Engine, error) {
	cfg := c.Config()
	logger := c.Logger().With("module", ModuleName)

	if !cfg.GuardrailsEnabled {
		return guardrailsapp.NewNoopGuardrails(), nil
	}

	db := core.MustGet[*pg.DB](c, database.GuardrailConn)
	if db == nil {
		return nil, fmt.Errorf("guardrails database connection is required")
	}

	repo := guardrailsinfra.NewPGRuleRepository(db.Pool())
	sink := guardrailsadapters.NewLoggerSink(logger)

	return guardrailsapp.NewSimpleEngine(repo, logger, sink), nil
}
