package remote

import (
	"context"
	"fmt"
	"strings"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/grpcx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	"github.com/jeanmolossi/maigo/pkg/maigo"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/jeanmolossi/maigo/pkg/maigo/header"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

const ModuleName = "gateway:remote"

type module struct{}

// NewModule wires HTTP clients that talk to standalone bounded-context services.
func NewModule() core.Module { return &module{} }

func (m *module) Name() string                                  { return ModuleName }
func (m *module) Weight() int                                   { return 2 }
func (m *module) Routes(_ *echo.Group, _ *core.Container) error { return nil }

func (m *module) Provide(ctx context.Context, c *core.Container) error {
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

	tenantStore, tenantConn, err := buildTenantStore(ctx, conf, newClient)
	if err != nil {
		return err
	}

	c.Set(core.TenantStoreModule, tenantStore)
	c.Set(core.GuardrailsModule, newGuardrailsClient(newClient(conf.Services.GuardURL)))
	c.Set(core.UsagePublisherModule, newUsageClient(newClient(conf.Services.ObsURL)))
	c.Set(core.AuditPublisherModule, newAuditClient(newClient(conf.Services.ObsURL)))

	if tenantConn != nil {
		c.Set(core.GRPCClientModule("gateway:tenant"), tenantConn)
	}

	return nil
}

func (m *module) Start(_ context.Context, c *core.Container) (func(context.Context) error, error) {
	val, ok := c.Get(core.GRPCClientModule("gateway:tenant"))
	if !ok {
		return nil, nil
	}

	conn, ok := val.(*grpc.ClientConn)
	if !ok || conn == nil {
		return nil, nil
	}

	return func(ctx context.Context) error {
		return conn.Close()
	}, nil
}

func buildTenantStore(
	ctx context.Context,
	conf *toolkitconfig.Config,
	newClient func(string) maigocontracts.ClientHTTPMethods,
) (tenantStore, *grpc.ClientConn, error) {
	if conf.Services.UseGRPC {
		conn, err := grpcx.Dial(ctx, grpcx.ClientConfig{
			Address:     conf.Services.TenantGRPCEndpoint,
			UseTLS:      conf.Services.GRPCTLS,
			CACertFile:  conf.Services.CACertFile,
			ServerName:  conf.Services.GRPCServerName,
			DialTimeout: conf.Services.RequestTimeout,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("connect tenant grpc: %w", err)
		}

		return newTenantGRPCClient(conn), conn, nil
	}

	return newTenantClient(newClient(conf.Services.TenantURL)), nil, nil
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
