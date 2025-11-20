package aiproxy

import (
	"context"
	"log/slog"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/labstack/echo/v4"
)

// ModuleName is the identifier of this module.
const ModuleName = "ai-proxy"

// module implements core.Module and is responsible for the AI proxy data-plane.
type module struct{}

// NewModule creates a new AI proxy module.
func NewModule() core.Module {
	return &module{}
}

// Name returns the module name for logging and debugging.
func (m *module) Name() string {
	return ModuleName
}

// Weight controls the initialization order; lower values are started earlier.
// AI proxy depends on core infra (config, logging, etc.), so it can be a low, but not the lowest, value.
func (m *module) Weight() int {
	return 10
}

// Migrations runs all database migrations required by this module.
// For the MVP, this is a no-op and can be wired later to a real migration engine.
func (m *module) Migrations(ctx context.Context, c *core.Container) error {
	// TODO: plug real migrations here (e.g., goose, atlas, migrate).
	log := c.Logger()
	log.Info("aiproxy: migrations skipped (MVP)")
	return nil
}

// Provide registers all dependencies required by this module into the container.
// This includes tenant store, rate limiter, semantic cache, router, guardrails, usage/audit publishers, etc.
func (m *module) Provide(c *core.Container) error {
	log := c.Logger()
	cfg := c.Config()

	log.Info("aiproxy: providing dependencies", slog.String("env", cfg.Environment.String()))

	// Wire all dependencies needed by HTTP handlers.
	deps, err := buildDependencies(c)
	if err != nil {
		return err
	}

	// Register a strong-typed struct or individual objects into the container if needed.
	c.Set(DepsKey, deps)

	return nil
}

// Routes registers HTTP routes handled by this module.
// It wires the AI proxy HTTP endpoints to Echo using the previously built dependencies.
func (m *module) Routes(e *echo.Echo, c *core.Container) error {
	log := c.Logger()
	log.Info("aiproxy: registering routes")

	deps := MustDepsFromContainer(c)

	registerRoutes(e, deps)

	return nil
}

// Start starts background workers if this module needs any.
// For the AI proxy MVP, this function is a no-op and returns nil stop function.
func (m *module) Start(ctx context.Context, c *core.Container) (func(context.Context) error, error) {
	// Example: start async usage flushers, cache warmers, etc. in the future.
	c.Logger().Info("aiproxy: start (no background workers for MVP)")
	return nil, nil
}
