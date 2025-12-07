package remote

import (
	"context"
	"encoding/json"
	"fmt"

	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	tenantpb "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tenantGRPCClient struct {
	client tenantpb.TenantServiceClient
}

func newTenantGRPCClient(conn *grpc.ClientConn) tenantStore {
	return &tenantGRPCClient{client: tenantpb.NewTenantServiceClient(conn)}
}

func (t *tenantGRPCClient) FindByAPIKey(ctx context.Context, apiKey string) (*pkgtenant.TenantConfig, error) {
	resp, err := t.client.GetByAPIKey(ctx, &tenantpb.GetTenantByAPIKeyRequest{ApiKey: apiKey})
	if err != nil {
		return nil, mapTenantError(err)
	}

	return tenantFromProto(resp.GetTenant())
}

func (t *tenantGRPCClient) FindByID(ctx context.Context, tenantID string) (*pkgtenant.TenantConfig, error) {
	resp, err := t.client.GetByID(ctx, &tenantpb.GetTenantByIDRequest{TenantId: tenantID})
	if err != nil {
		return nil, mapTenantError(err)
	}

	return tenantFromProto(resp.GetTenant())
}

func (t *tenantGRPCClient) RevokeExpired(context.Context) (int64, error) {
	return 0, nil
}

func tenantFromProto(src *tenantpb.Tenant) (*pkgtenant.TenantConfig, error) {
	if src == nil {
		return nil, fmt.Errorf("empty tenant payload")
	}

	cfg := &pkgtenant.TenantConfig{
		ID:                  src.GetId(),
		Code:                src.GetCode(),
		Name:                src.GetName(),
		Status:              src.GetStatus(),
		PlanCode:            src.GetPlanCode(),
		MaxTokensMonth:      src.GetMaxTokensMonth(),
		MaxRequestsMinute:   src.GetMaxRequestsMinute(),
		DefaultModel:        src.GetDefaultModel(),
		EnableSemanticCache: src.GetEnableSemanticCache(),
		CacheTTLSecs:        src.GetCacheTtlSecs(),
		MaxPromptTokens:     src.GetMaxPromptTokens(),
		MaxCompletionTokens: src.GetMaxCompletionTokens(),
		PolicyConfigRaw:     src.GetPolicyConfig(),
	}

	if len(cfg.PolicyConfigRaw) > 0 {
		var parsed pkgtenant.PolicyConfig
		if err := json.Unmarshal(cfg.PolicyConfigRaw, &parsed); err == nil {
			cfg.ParsedPolicyConfig = &parsed
		}
	}

	return cfg, nil
}

func mapTenantError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.InvalidArgument:
		return pkgtenant.ErrInvalidAPIKey
	case codes.NotFound:
		return pkgtenant.ErrNotFound
	case codes.FailedPrecondition, codes.PermissionDenied:
		return pkgtenant.ErrInactiveTenant
	default:
		return err
	}
}
