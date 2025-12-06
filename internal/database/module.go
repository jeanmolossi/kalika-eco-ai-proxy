package database

import (
	"context"
	"fmt"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/labstack/echo/v4"
)

// Module identifiers and connection keys for each bounded context.
const (
	ModuleName        = "database"
	GatewayConn       = ModuleName + ":gateway"
	TenantConn        = ModuleName + ":tenant"
	GuardrailConn     = ModuleName + ":guardrail"
	ObservabilityConn = ModuleName + ":observability"
)

// ModuleOptions describes how to instantiate a database module bound to a specific config slice.
type ModuleOptions struct {
	Name     string
	ConnKey  string
	Selector func(*config.Config) config.Postgres
}

type module struct {
	name     string
	connKey  string
	selector func(*config.Config) config.Postgres
}

// NewModule builds a module that exposes a dedicated Postgres connection using the provided selector.
func NewModule(opt ModuleOptions) core.Module {
	if opt.Name == "" {
		opt.Name = ModuleName
	}

	if opt.ConnKey == "" {
		panic("database module requires ConnKey")
	}

	return &module{
		name:     opt.Name,
		connKey:  opt.ConnKey,
		selector: opt.Selector,
	}
}

// NewGatewayModule opens the gateway database connection.
func NewGatewayModule() core.Module {
	return NewModule(ModuleOptions{
		Name:    ModuleName + "/gateway",
		ConnKey: GatewayConn,
		Selector: func(cfg *config.Config) config.Postgres {
			return cfg.GatewayDB
		},
	})
}

// NewTenantModule opens the tenant database connection.
func NewTenantModule() core.Module {
	return NewModule(ModuleOptions{
		Name:    ModuleName + "/tenant",
		ConnKey: TenantConn,
		Selector: func(cfg *config.Config) config.Postgres {
			return cfg.TenantDB
		},
	})
}

// NewGuardrailModule opens the guardrail database connection.
func NewGuardrailModule() core.Module {
	return NewModule(ModuleOptions{
		Name:    ModuleName + "/guardrail",
		ConnKey: GuardrailConn,
		Selector: func(cfg *config.Config) config.Postgres {
			return cfg.GuardDB
		},
	})
}

// NewObservabilityModule opens the observability database connection.
func NewObservabilityModule() core.Module {
	return NewModule(ModuleOptions{
		Name:    ModuleName + "/observability",
		ConnKey: ObservabilityConn,
		Selector: func(cfg *config.Config) config.Postgres {
			return cfg.ObserveDB
		},
	})
}

// Name returns the module name for logging and debugging.
func (m *module) Name() string { return m.name }

// Weight ensures the connection is ready before dependent modules.
func (m *module) Weight() int { return 1 }

// Provide implements core.Module.
func (m *module) Provide(ctx context.Context, c *core.Container) error {
	if m.selector == nil {
		return fmt.Errorf("%s: missing postgres selector", m.name)
	}

	dbCfg := m.selector(c.Config())
	if dbCfg.DSN == "" && dbCfg.Database.Database == "" {
		return fmt.Errorf("%s: postgres configuration is empty", m.name)
	}

	pgdb, err := pg.Open(ctx, dbCfg)
	if err != nil {
		return err
	}

	c.Set(m.connKey, pgdb)

	return nil
}

// Routes implements core.Module.
func (m *module) Routes(g *echo.Group, c *core.Container) error { return nil }

// Start implements core.Module.
func (m *module) Start(ctx context.Context, c *core.Container) (stop func(context.Context) error, err error) {
	return stop, err
}

// Migrations implements core.Module.
func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
