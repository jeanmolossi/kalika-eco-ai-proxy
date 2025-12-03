package remote

import (
	"context"
	"strings"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/labstack/echo/v4"
)

const ModuleName = "guardrails:remote"

type module struct{}

// NewModule registers the HTTP clients guardrails needs to talk to other services.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 1 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	conf := c.Config()

	newClient := func(baseURL string) maigocontracts.ClientHTTPMethods {
		return maigo.NewClient(strings.TrimSuffix(baseURL, "/")).Config().
			SetTimeout(10 * time.Second).
			Build()
	}

	c.Set(core.TenantStoreModule, newTenantClient(newClient(conf.Services.TenantURL)))

	return nil
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, _ *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}
