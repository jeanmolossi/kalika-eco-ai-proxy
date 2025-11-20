package tenant

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresStore struct {
	db *pgxpool.Pool
}

// FindByAPIKey implements Store.
func (p *postgresStore) FindByAPIKey(ctx context.Context, apiKey string) (*TenantConfig, error) {
	prefix, err := extractPrefix(apiKey)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256([]byte(strings.TrimSpace(strings.TrimPrefix(apiKey, "Bearer "))))
	secretHash := sum[:]

	const query = `
        SELECT
            t.id,
            t.code,
            t.name,
            t.status::text,
            t.plan_code,
            t.max_tokens_month,
            t.max_requests_minute,
            p.default_model,
            p.enable_semantic_cache,
            p.cache_ttl_seconds,
            p.max_prompt_tokens,
            p.max_completion_tokens,
            COALESCE(p.config, '{}'::jsonb) AS policy_config
        FROM apx.tenant_api_keys k
        JOIN apx.tenants t ON t.id = k.tenant_id
        LEFT JOIN apx.tenant_policies p ON p.tenant_id = t.id
            AND p.is_active = TRUE
        WHERE
            k.prefix = $1
            AND k.secret_hash = $2
            AND k.status = 'active'
            AND (k.expires_at IS NULL OR k.expires_at > now())
        LIMIT 1;`

	row := p.db.QueryRow(ctx, query, prefix, secretHash)

	var (
		cfg            TenantConfig
		policyJSON     []byte
		status         string
		maxTokensMonth int64
		maxReqMin      int32
	)

	err = row.Scan(
		&cfg.ID,
		&cfg.Code,
		&cfg.Name,
		&status,
		&cfg.PlanCode,
		&maxTokensMonth,
		&maxReqMin,
		&cfg.DefaultModel,
		&cfg.EnableSemanticCache,
		&cfg.CacheTTLSecs,
		&cfg.MaxPromptTokens,
		&cfg.MaxCompletionTokens,
		&policyJSON,
	)
	if err != nil {
		return nil, ErrNotFound
	}

	cfg.Status = status
	cfg.MaxTokensMonth = maxTokensMonth
	cfg.MaxRequestsMinute = int64(maxReqMin)
	cfg.PolicyConfigRaw = policyJSON

	if len(policyJSON) > 0 && string(policyJSON) != "{}" {
		var pc PolicyConfig
		if err := json.Unmarshal(policyJSON, &pc); err == nil {
			cfg.ParsedPolicyConfig = &pc
		}
	}

	if cfg.Status != "active" && cfg.Status != "trialing" {
		return nil, ErrInactiveTenant
	}

	return &cfg, nil
}

// FindByID implements Store.
func (p *postgresStore) FindByID(ctx context.Context, tenantID string) (*TenantConfig, error) {
	const query = `
        SELECT
            t.id,
            t.code,
            t.name,
            t.status::text,
            t.plan_code,
            t.max_tokens_month,
            t.max_requests_minute,
            p.default_model,
            p.enable_semantic_cache,
            p.cache_ttl_seconds,
            p.max_prompt_tokens,
            p.max_completion_tokens,
            COALESCE(p.config, '{}'::jsonb) AS policy_config
        FROM apx.tenants t
        LEFT JOIN apx.tenant_policies p ON p.tenant_id = t.id AND p.is_active = TRUE
        WHERE t.id = $1
        LIMIT 1;`

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	row := p.db.QueryRow(ctx, query, tenantID)

	var (
		cfg            TenantConfig
		policyJSON     []byte
		status         string
		maxTokensMonth int64
		maxReqMin      int32
	)

	err := row.Scan(
		&cfg.ID,
		&cfg.Code,
		&cfg.Name,
		&status,
		&cfg.PlanCode,
		&maxTokensMonth,
		&maxReqMin,
		&cfg.DefaultModel,
		&cfg.EnableSemanticCache,
		&cfg.CacheTTLSecs,
		&cfg.MaxPromptTokens,
		&cfg.MaxCompletionTokens,
		&policyJSON,
	)
	if err != nil {
		return nil, ErrNotFound
	}

	cfg.Status = status
	cfg.MaxTokensMonth = maxTokensMonth
	cfg.MaxRequestsMinute = int64(maxReqMin)
	cfg.PolicyConfigRaw = policyJSON

	if len(policyJSON) > 0 && string(policyJSON) != "{}" {
		var pc PolicyConfig
		if err := json.Unmarshal(policyJSON, &pc); err == nil {
			cfg.ParsedPolicyConfig = &pc
		}
	}

	if cfg.Status != "active" && cfg.Status != "trialing" {
		return nil, ErrInactiveTenant
	}

	return &cfg, nil
}

func extractPrefix(apiKey string) (string, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", ErrInvalidAPIKey
	}

	// allows "Bearer xxx"
	if strings.HasPrefix(strings.ToLower(apiKey), "bearer ") {
		apiKey = strings.TrimSpace(apiKey[7:])
	}

	idx := strings.LastIndex(apiKey, "_")
	if idx <= 0 {
		return "", fmt.Errorf("%w: missing underscore separator", ErrInvalidAPIKey)
	}

	prefix := apiKey[:idx]
	return prefix, nil
}

func NewPostgresStore(pool *pgxpool.Pool) Store {
	return &postgresStore{db: pool}
}
