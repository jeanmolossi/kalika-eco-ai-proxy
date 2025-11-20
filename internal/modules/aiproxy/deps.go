package aiproxy

import (
	"fmt"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/audit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/cache"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/router"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/usage"
)

// DepsKey is the container key used to store AI proxy dependencies.
const DepsKey = "aiproxy:deps"

// Deps groups all dependencies required by the AI proxy HTTP layer.
type Deps struct {
	TenantStore tenant.Store
	Router      router.Router
	Cache       cache.SemanticCache
	Guardrails  guardrails.Guardrails
	UsagePub    usage.Publisher
	AuditPub    audit.Publisher
	Limiter     ratelimit.Limiter
	LLMClient   llm.Client
}

// MustDepsFromContainer retrieves Deps from container or panics.
func MustDepsFromContainer(c *core.Container) Deps {
	v := c.MustGet(DepsKey)

	deps, ok := v.(Deps)
	if !ok {
		panic("aiproxy: invalid deps type stored in container")
	}

	return deps
}

// buildDependencies constructs the Deps struct with minimal stub implementations.
// This is meant to be the MVP wiring and can be replaced with real implementations over time.
func buildDependencies(c *core.Container) (Deps, error) {
	// For the MVP, create very simple noop/stub implementations.
	// You can later replace these with real DB/Redis/Kafka/etc. integrations.
	tenantStore := tenant.NewInMemoryStore()
	limiter := ratelimit.NewNoopLimiter()
	semCache := cache.NewNoopSemanticCache()
	guard := guardrails.NewNoopGuardrails()
	usagePub := usage.NewLogPublisher(c.Logger())
	auditPub := audit.NewLogPublisher(c.Logger())
	llmClient := llm.NewStubClient() // or a real OpenAI/Anthropic client

	rt := router.NewSimpleRouter(llmClient)

	deps := Deps{
		TenantStore: tenantStore,
		Router:      rt,
		Cache:       semCache,
		Guardrails:  guard,
		UsagePub:    usagePub,
		AuditPub:    auditPub,
		Limiter:     limiter,
		LLMClient:   llmClient,
	}

	// Sanity check before returning.
	if deps.TenantStore == nil || deps.Router == nil {
		return Deps{}, fmt.Errorf("aiproxy: invalid deps, some fields are nil")
	}

	return deps, nil
}
