package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	"github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/header"
	"github.com/labstack/echo/v4"
)

const ModuleName = "gateway:remote"

type module struct{}

// NewModule wires HTTP clients that talk to standalone bounded-context services.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 2 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(_ context.Context, c *core.Container) error {
	conf := c.Config()

	clientHTTP, err := httpx.NewServiceHTTPClient(httpx.ServiceClientOptions{
		Timeout:    conf.Services.RequestTimeout,
		MaxRetries: conf.Services.MaxRetries,
		Breaker: httpx.CircuitBreakerConfig{
			FailureThreshold: int(conf.Services.CircuitFailures),
			RecoveryWindow:   conf.Services.CircuitReset,
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
	c.Set(core.GuardrailsModule, newGuardrailsClient(newClient(conf.Services.GuardURL)))
	c.Set(core.UsagePublisherModule, newUsageClient(newClient(conf.Services.ObsURL)))
	c.Set(core.AuditPublisherModule, newAuditClient(newClient(conf.Services.ObsURL)))

	return nil
}

func (m *module) Start(_ context.Context, _ *core.Container) (func(context.Context) error, error) {
	return nil, nil
}

func (m *module) Migrations(ctx context.Context, c *core.Container) ([]core.MigrationFile, error) {
	return make([]core.MigrationFile, 0), nil
}

type guardrailsEngine interface {
	EvaluateInput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error)
	EvaluateOutput(ctx context.Context, gx guardrails.Context) (guardrails.Decision, error)
}

type tenantStore interface {
	FindByAPIKey(ctx context.Context, apiKey string) (*pkgtenant.TenantConfig, error)
	FindByID(ctx context.Context, tenantID string) (*pkgtenant.TenantConfig, error)
	RevokeExpired(ctx context.Context) (int64, error)
}

type usagePublisher interface {
	Publish(ctx context.Context, ev observability.UsageEvent) error
}

type auditPublisher interface {
	Publish(ctx context.Context, ev observability.AuditEvent) error
}
