package guardrails

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRuleRepository struct {
	pool *pgxpool.Pool
}

func NewPGRuleRepository(pool *pgxpool.Pool) RuleRepository {
	return &pgRuleRepository{pool: pool}
}

// ListRulesForTenantPhase implements RuleRepository.
func (p *pgRuleRepository) ListRulesForTenantPhase(ctx context.Context, tenantID string, phase Phase) ([]Rule, error) {
	const query = `
        SELECT 
            id,
            tenant_id,
            name,
            kind,
            is_active,
            priority,
            config
        FROM apx.guardrail_rules
        WHERE tenant_id = $1
            AND is_active = true
            AND (config->>'phase' = $2 OR config->>'phase' IS NULL)
        ORDER BY priority ASC, created_at ASC;`

	rows, err := p.pool.Query(ctx, query, tenantID, phase)
	if err != nil {
		return nil, fmt.Errorf("query guardrail rules: %w", err)
	}
	defer rows.Close()

	out := make([]Rule, 0)

	for rows.Next() {
		var (
			rule   Rule
			rawCfg []byte
		)

		if err := rows.Scan(
			&rule.ID,
			&rule.TenantID,
			&rule.Name,
			&rule.Kind,
			&rule.IsActive,
			&rule.Priority,
			&rawCfg,
		); err != nil {
			return nil, fmt.Errorf("scan guardrail rule: %w", err)
		}

		if len(rawCfg) == 0 {
			rule.Config = RuleConfig{}
		} else {
			if err := json.Unmarshal(rawCfg, &rule.Config); err != nil {
				continue
			}
		}

		if rule.Config.Phase == "" {
			rule.Config.Phase = PhaseInput
		}

		out = append(out, rule)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows err: %w", rows.Err())
	}

	return out, nil
}
