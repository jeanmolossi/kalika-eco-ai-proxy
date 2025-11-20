package core

import (
	"log/slog"
	"sync"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/config"
)

// Container is a simple dependency registry (like a controlled service locator).
type Container struct {
	byKey sync.Map
}

func NewContainer() *Container {
	return &Container{byKey: sync.Map{}}
}

func (c *Container) Set(key string, v any) {
	c.byKey.Store(key, v)
}

func (c *Container) Get(key string) (any, bool) {
	return c.byKey.Load(key)
}

// MustGet returns the value or panics if it doesn't exist.
// Useful for mandatory dependencies (Config, Logger, etc.).
func (c *Container) MustGet(key string) any {
	v, ok := c.Get(key)
	if !ok {
		panic("core.Container: missing dependency for key: " + key)
	}

	return v
}

// MustGet T is the generic type-safe version.
func MustGet[T any](c *Container, key string) T {
	v, ok := c.Get(key)
	if !ok {
		panic("core.Container: missing dependency for key: " + key)
	}

	casted, ok := v.(T)
	if !ok {
		panic("core.Container: invalid type assertion for key: " + key)
	}

	return casted
}

// Specific helpers for standard dependencies:

func (c *Container) Config() *config.Config {
	return MustGet[*config.Config](c, ConfigModule)
}

func (c *Container) Logger() *slog.Logger {
	return MustGet[*slog.Logger](c, LoggerModule)
}
