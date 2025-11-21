package app

import (
	"context"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/tokenizer"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/usage"
)

// ===========================================================================
//
// IO
//
// ===========================================================================

// ChatInput represents the high-level input for a chat completion call.
type ChatInput struct {
	Tenant   tenant.TenantConfig
	UserID   string
	Metadata map[string]string
	Request  llm.ChatRequest
}

// ChatOutput is a type alias for the low-level LLM chat response.
type ChatOutput = llm.ChatResponse

// EmbeddingsInput represents the input for an embeddings request.
type EmbeddingsInput struct {
	Tenant   tenant.TenantConfig
	UserID   string
	Metadata map[string]string
	Request  llm.EmbedRequest
}

// EmbeddingsOutput is a type alias for the embeddings response.
type EmbeddingsOutput = llm.EmbedResponse

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

type TokenCounter = tokenizer.TokenCounter

// TokenLimiter is a narrow interface for rate limiting required by the service.
type TokenLimiter = ratelimit.Limiter

// SemanticCache is the minimal interface the service needs for semantic caching.
type SemanticCache interface {
	LookupChat(ctx context.Context, tenantID string, req llm.ChatRequest) (*llm.ChatResponse, bool, error)
	StoreChat(ctx context.Context, tenantID string, req llm.ChatRequest, resp llm.ChatResponse) error
}

// ChatGuardrails represents the subset of guardrails features used by the chat flow.
type ChatGuardrails interface {
	PreProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatRequest, error)
	PostProcessChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest, resp llm.ChatResponse) (llm.ChatResponse, error)
}

// ChatRouter is the routing interface for chat and embeddings.
type ChatRouter interface {
	RouteChat(ctx context.Context, t tenant.TenantConfig, req llm.ChatRequest) (llm.ChatResponse, error)
	RouteEmbed(ctx context.Context, t tenant.TenantConfig, req llm.EmbedRequest) (llm.EmbedResponse, error)
}

// UsagePublisher is the minimal usage event publisher.
type UsagePublisher interface {
	Publish(ctx context.Context, ev usage.Event) error
}

// AuditPublisher is the minimal audit event publisher.
type AuditPublisher interface {
	Publish(ctx context.Context, ev audit.Event) error
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
