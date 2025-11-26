package app

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	pkgguardrails "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/guardrails"
	pkgllm "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/llm"
	observability "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/observability"
	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
	pkgtokenizer "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tokenizer"
)

// ===========================================================================
//
// IO
//
// ===========================================================================

// ChatInput represents the high-level input for a chat completion call.
type ChatInput struct {
	Tenant   pkgtenant.TenantConfig
	APIKey   string
	UserID   string
	Metadata map[string]string
	Request  pkgllm.ChatRequest
}

// ChatOutput is a type alias for the low-level LLM chat response.
type ChatOutput = pkgllm.ChatResponse

// EmbeddingsInput represents the input for an embeddings request.
type EmbeddingsInput struct {
	Tenant   pkgtenant.TenantConfig
	UserID   string
	Metadata map[string]string
	Request  pkgllm.EmbedRequest
}

// EmbeddingsOutput is a type alias for the embeddings response.
type EmbeddingsOutput = pkgllm.EmbedResponse

// ===========================================================================
//
// UseCases
//
// ===========================================================================

// ChatUseCase defines the minimal interface for the chat flow.
type ChatUseCase interface {
	Chat(ctx context.Context, in ChatInput) (ChatOutput, error)
}

// EmbeddingsUseCase defines the minimal interface for embeddings.
type EmbeddingsUseCase interface {
	Embeddings(ctx context.Context, in EmbeddingsInput) (EmbeddingsOutput, error)
}

// Service implements both ChatUseCase and EmbeddingsUseCase.
var (
	_ ChatUseCase       = (*Service)(nil)
	_ EmbeddingsUseCase = (*Service)(nil)
)

// ===========================================================================
//
// Service deps
//
// ===========================================================================

type TokenCounter = pkgtokenizer.TokenCounter

// TokenLimiter is a narrow interface for rate limiting required by the service.
type TokenLimiter = ratelimit.Limiter

// SemanticCache is the minimal interface the service needs for semantic caching.
type SemanticCache interface {
	LookupChat(ctx context.Context, tenantID string, req pkgllm.ChatRequest) (*pkgllm.ChatResponse, bool, error)
	StoreChat(ctx context.Context, tenantID string, req pkgllm.ChatRequest, resp pkgllm.ChatResponse) error
}

// ChatGuardrails represents the subset of guardrails features used by the chat flow.
type ChatGuardrails interface {
	EvaluateInput(ctx context.Context, gx pkgguardrails.Context) (pkgguardrails.Decision, error)
	EvaluateOutput(ctx context.Context, gx pkgguardrails.Context) (pkgguardrails.Decision, error)
}

// ChatRouter is the routing interface for chat and embeddings.
type ChatRouter interface {
	RouteChat(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.ChatRequest) (pkgllm.ChatResponse, error)
	RouteEmbed(ctx context.Context, t pkgtenant.TenantConfig, req pkgllm.EmbedRequest) (pkgllm.EmbedResponse, error)
}

// UsagePublisher is the minimal usage event publisher.
type UsagePublisher interface {
	Publish(ctx context.Context, ev observability.UsageEvent) error
}

// AuditPublisher is the minimal audit event publisher.
type AuditPublisher interface {
	Publish(ctx context.Context, ev observability.AuditEvent) error
}

// Service orchestrates the AI proxy flow using small, segregated interfaces.
type Service struct {
	limiter    TokenLimiter
	cache      SemanticCache
	guardrails ChatGuardrails
	router     ChatRouter
	usagePub   UsagePublisher
	auditPub   AuditPublisher
	tokenizr   TokenCounter
}

// NewService creates a new Service instance.
func NewService(
	limiter TokenLimiter,
	cache SemanticCache,
	guard ChatGuardrails,
	rt ChatRouter,
	usagePub UsagePublisher,
	auditPub AuditPublisher,
	tokenizr TokenCounter,
) *Service {
	return &Service{
		limiter:    limiter,
		cache:      cache,
		guardrails: guard,
		router:     rt,
		usagePub:   usagePub,
		auditPub:   auditPub,
		tokenizr:   tokenizr,
	}
}
