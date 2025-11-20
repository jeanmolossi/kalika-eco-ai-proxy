package tenant

import (
	"fmt"
	"sync"
)

// DefaultDevAPIKey is the default API key used by the in-memory stub store.
// It exists only for local development and quick testing.
const DefaultDevAPIKey = "dev-api-key"

// InMemoryStore is a simple in-memory implementation of the tenant store.
// It is intended for local development and tests only.
type InMemoryStore struct {
	mu    sync.RWMutex
	byID  map[string]*TenantConfig
	byKey map[string]*TenantConfig
}

// NewInMemoryStore creates a new in-memory tenant store with an optional
// pre-populated default tenant for local development.
func NewInMemoryStore() *InMemoryStore {
	s := &InMemoryStore{
		byID:  make(map[string]*TenantConfig),
		byKey: make(map[string]*TenantConfig),
	}

	// Seed a default tenant for quick manual testing.
	s.AddTenant(&TenantConfig{
		ID:                  "tnt_dev",
		Name:                "Dev Tenant",
		Plan:                "dev",
		MaxTokensMonth:      1_000_000,
		ModelsAllowed:       []string{"stub-model"},
		DefaultModel:        "stub-model",
		EnableSemanticCache: false,
		MaxRequestsMinute:   120,
	}, DefaultDevAPIKey)

	return s
}

// AddTenant registers a tenant configuration with a given API key.
func (s *InMemoryStore) AddTenant(cfg *TenantConfig, apiKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byID[cfg.ID] = cfg
	s.byKey[apiKey] = cfg
}

// FindByAPIKey returns the tenant configuration associated with the given API key.
func (s *InMemoryStore) FindByAPIKey(apiKey string) (*TenantConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cfg, ok := s.byKey[apiKey]; ok {
		return cfg, nil
	}

	return nil, fmt.Errorf("tenant not found for api key")
}

// FindByID returns the tenant configuration for a given tenant ID.
func (s *InMemoryStore) FindByID(id string) (*TenantConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if cfg, ok := s.byID[id]; ok {
		return cfg, nil
	}

	return nil, fmt.Errorf("tenant not found for id")
}
