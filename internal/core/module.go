package core

import (
	"context"

	"github.com/labstack/echo/v4"
)

// Module represents a logical unit of the application.
// E.g.: "ai-proxy", "admin", "billing", etc.
type Module interface {
	// Name returns the name of the module (for logs, errors, debugging).
	Name() string

	// Weight defines the initialization order.
	// The smaller the value, the earlier the module runs (migrations/provide/routes/start).
	Weight() int

	// Migrations is the point to retrieve database migrations, seeds, etc.
	// It can be a no-op if not needed.
	Migrations(ctx context.Context, c *Container) ([]MigrationFile, error)

	// Provide registers dependencies in the container (repos, services, use-cases).
	Provide(ctx context.Context, c *Container) error

	// Routes registers the HTTP routes of the module.
	// There is no need to start the server here, just register in e.
	Routes(e *echo.Echo, c *Container) error

	// Start initiates background routines, consumers, etc.
	// Returns a stop function for graceful shutdown.
	Start(ctx context.Context, c *Container) (stop func(context.Context) error, err error)
}
