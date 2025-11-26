package llm

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	pkgtenant "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/tenant"
)

// ProviderPool resolves the correct upstream client for a tenant/model pair.
// It builds clients lazily from tenant routing configuration while caching
// instances to reuse HTTP connections and metrics hooks.
type ProviderPool struct {
	defaultProvider ProviderSettings
	metrics         MetricsRecorder
	factory         StrategyFactory

	mu       sync.RWMutex
	clients  map[string]Client
	override map[string][]ProviderSettings
}

// NewProviderPool builds a pool with a global default provider.
func NewProviderPool(defaults ProviderSettings, metrics MetricsRecorder) *ProviderPool {
	if metrics == nil {
		metrics = NoopMetrics{}
	}

	return &ProviderPool{
		defaultProvider: defaults,
		metrics:         metrics,
		factory:         NewStrategyFactory(metrics),
		clients:         make(map[string]Client),
		override:        make(map[string][]ProviderSettings),
	}
}

// ClientFor returns an LLM client that supports the given model for the tenant.
func (p *ProviderPool) ClientFor(ctx context.Context, t pkgtenant.TenantConfig, model string) (Client, error) {
	_ = ctx
	providers := p.resolveProviders(t)

	for _, cfg := range providers {
		if supportsModel(cfg, model) {
			return p.clientForConfig(cfg)
		}
	}

	return nil, errors.New("llm: no provider available for model")
}

func (p *ProviderPool) resolveProviders(t pkgtenant.TenantConfig) []ProviderSettings {
	if t.ParsedPolicyConfig == nil || t.ParsedPolicyConfig.Routing == nil || len(t.ParsedPolicyConfig.Routing.Providers) == 0 {
		return []ProviderSettings{p.defaultProvider}
	}

	return p.cachedTenantProviders(t.ID, t.ParsedPolicyConfig.Routing.Providers)
}

func (p *ProviderPool) cachedTenantProviders(tenantID string, defs []pkgtenant.ProviderDefinition) []ProviderSettings {
	p.mu.RLock()
	cfgs, ok := p.override[tenantID]
	p.mu.RUnlock()

	if ok {
		return cfgs
	}

	built := make([]ProviderSettings, 0, len(defs))
	for _, d := range defs {
		built = append(built, ProviderSettings{
			Name:            d.Name,
			BaseURL:         d.BaseURL,
			APIKey:          d.APIKey,
			RequestTimeout:  timeoutOrDefault(d.RequestTimeoutMS),
			MaxRetries:      d.MaxRetries,
			EnableStreaming: d.EnableStreaming,
			ChatModels:      d.ChatModels,
			EmbedModels:     d.EmbedModels,
		})
	}

	p.mu.Lock()
	p.override[tenantID] = built
	p.mu.Unlock()

	return built
}

func (p *ProviderPool) clientForConfig(cfg ProviderSettings) (Client, error) {
	key := cfg.Name + "@" + cfg.BaseURL

	p.mu.RLock()
	client, ok := p.clients[key]
	p.mu.RUnlock()

	if ok {
		return client, nil
	}

	client, err := p.factory.Build(cfg)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	p.clients[key] = client
	p.mu.Unlock()

	return client, nil
}

func supportsModel(cfg ProviderSettings, model string) bool {
	if len(cfg.ChatModels) == 0 && len(cfg.EmbedModels) == 0 {
		return true
	}

	return slices.Contains(cfg.ChatModels, model) || slices.Contains(cfg.EmbedModels, model)
}

func timeoutOrDefault(ms int) time.Duration {
	if ms <= 0 {
		return 20 * time.Second
	}

	return time.Duration(ms) * time.Millisecond
}
