package grpcadapter

import (
	"context"
	"errors"

	tenantapp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/app"
	tenantpb "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	tenantpb.UnimplementedTenantServiceServer
	store tenantapp.Store
}

func NewServer(store tenantapp.Store) *Server {
	return &Server{store: store}
}

func (s *Server) GetByAPIKey(ctx context.Context, req *tenantpb.GetTenantByAPIKeyRequest) (*tenantpb.GetTenantResponse, error) {
	if req == nil || req.ApiKey == "" {
		return nil, status.Error(codes.InvalidArgument, "api_key is required")
	}

	tenant, err := s.store.FindByAPIKey(ctx, req.ApiKey)
	if err != nil {
		return nil, mapError(err)
	}

	return &tenantpb.GetTenantResponse{Tenant: toProto(tenant)}, nil
}

func (s *Server) GetByID(ctx context.Context, req *tenantpb.GetTenantByIDRequest) (*tenantpb.GetTenantResponse, error) {
	if req == nil || req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	tenant, err := s.store.FindByID(ctx, req.TenantId)
	if err != nil {
		return nil, mapError(err)
	}

	return &tenantpb.GetTenantResponse{Tenant: toProto(tenant)}, nil
}

func toProto(cfg *tenantapp.TenantConfig) *tenantpb.Tenant {
	if cfg == nil {
		return nil
	}

	return &tenantpb.Tenant{
		Id:                  cfg.ID,
		Code:                cfg.Code,
		Name:                cfg.Name,
		Status:              cfg.Status,
		PlanCode:            cfg.PlanCode,
		MaxTokensMonth:      cfg.MaxTokensMonth,
		MaxRequestsMinute:   cfg.MaxRequestsMinute,
		DefaultModel:        cfg.DefaultModel,
		EnableSemanticCache: cfg.EnableSemanticCache,
		CacheTtlSecs:        cfg.CacheTTLSecs,
		MaxPromptTokens:     cfg.MaxPromptTokens,
		MaxCompletionTokens: cfg.MaxCompletionTokens,
		PolicyConfig:        cfg.PolicyConfigRaw,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, tenantapp.ErrInvalidAPIKey):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, tenantapp.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, tenantapp.ErrInactiveTenant):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
