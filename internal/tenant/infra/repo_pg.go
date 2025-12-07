package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	tenantapp "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/app"
)

type postgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) tenantapp.Store {
	return &postgresStore{pool: pool}
}

// FindByAPIKey retrieves tenant configuration by API key.
func (p *postgresStore) FindByAPIKey(ctx context.Context, apiKey string) (*tenantapp.TenantConfig, error) {
	if apiKey == "" {
		return nil, tenantapp.ErrInvalidAPIKey
	}

	const query = `
        SELECT
            t.id,
            t.code,
            t.name,
            t.status,
            t.plan_code,
            t.max_tokens_month,
            t.max_requests_minute,
            t.default_model,
            t.enable_semantic_cache,
            t.cache_ttl_secs,
            t.max_prompt_tokens,
            t.max_completion_tokens,
            t.policy_config
        FROM apx.tenants t
        JOIN apx.api_keys k ON k.tenant_id = t.id
        WHERE k.key = $1 AND k.expires_at > NOW();`

	row := p.pool.QueryRow(ctx, query, apiKey)

	return scanTenant(row)
}

func (p *postgresStore) FindByID(ctx context.Context, tenantID string) (*tenantapp.TenantConfig, error) {
	const query = `
        SELECT
            t.id,
            t.code,
            t.name,
            t.status,
            t.plan_code,
            t.max_tokens_month,
            t.max_requests_minute,
            t.default_model,
            t.enable_semantic_cache,
            t.cache_ttl_secs,
            t.max_prompt_tokens,
            t.max_completion_tokens,
            t.policy_config
        FROM apx.tenants t
        WHERE t.id = $1;`

	row := p.pool.QueryRow(ctx, query, tenantID)

	return scanTenant(row)
}

func (p *postgresStore) RevokeExpired(ctx context.Context) (int64, error) {
	const query = `
        UPDATE apx.api_keys SET expires_at = now() WHERE expires_at < now() RETURNING id;
        `

	cmd, err := p.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("revoke expired api keys: %w", err)
	}

	return cmd.RowsAffected(), nil
}

func scanTenant(row pgx.Row) (*tenantapp.TenantConfig, error) {
	var (
		tenant tenantapp.TenantConfig
		rawCfg []byte
	)

	if err := row.Scan(
		&tenant.ID,
		&tenant.Code,
		&tenant.Name,
		&tenant.Status,
		&tenant.PlanCode,
		&tenant.MaxTokensMonth,
		&tenant.MaxRequestsMinute,
		&tenant.DefaultModel,
		&tenant.EnableSemanticCache,
		&tenant.CacheTTLSecs,
		&tenant.MaxPromptTokens,
		&tenant.MaxCompletionTokens,
		&rawCfg,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, tenantapp.ErrNotFound
		}

		return nil, fmt.Errorf("scan tenant: %w", err)
	}

	if len(rawCfg) > 0 {
		tenant.PolicyConfigRaw = rawCfg

		var parsed tenantapp.PolicyConfig
		if err := json.Unmarshal(rawCfg, &parsed); err == nil {
			tenant.ParsedPolicyConfig = &parsed
		}
	}

	return &tenant, nil
}
