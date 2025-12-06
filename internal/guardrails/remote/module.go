package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	"github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/header"
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

	clientHTTP, err := httpx.NewServiceHTTPClient(httpx.ServiceClientOptions{
		Timeout:    conf.Services.RequestTimeout,
		MaxRetries: conf.Services.MaxRetries,
		Breaker: httpx.CircuitBreakerConfig{
			Failures:     uint32(conf.Services.CircuitFailures),
			ResetTimeout: conf.Services.CircuitReset,
			Interval:     conf.Services.CircuitInterval,
		},
		CACertFile: conf.Services.CACertFile,
	})
	if err != nil {
		return fmt.Errorf("create service http client: %w", err)
	}

	token := strings.TrimSpace(conf.Services.AuthToken)

	newClient := func(baseURL string) maigocontracts.ClientHTTPMethods {
		clientBuilder := maigo.NewClient(strings.TrimSuffix(baseURL, "/"))
		configBuilder := clientBuilder.Config()
		configBuilder.SetCustomHTTPClient(httpx.AsMaigoHTTPClient(clientHTTP))
		configBuilder.SetTimeout(conf.Services.RequestTimeout)

		if token != "" {
			clientBuilder.Header().Set(header.Authorization, "Bearer "+token)
		}

		return clientBuilder.Build()
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
