package guardrails

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type pgRuleRepository struct {
	pool *pgxpool.Pool
}

func NewPGRuleRepository(pool *pgxpool.Pool) RuleRepository {
	return &pgRuleRepository{pool: pool}
}

// ListRulesForTenant implements RuleRepository.
func (p *pgRuleRepository) ListRulesForTenant(ctx context.Context, tenantID string, phase Phase) ([]Rule, error) {
	const query = `
        SELECT 
            id,
            tenant_id,
            phase,
            kind,
            action,
            pattern,
            replacement,
            priority,
            enabled
        FROM apx.guardrail_rules
        WHERE tenant_id = $1
        AND phase = $2
        AND enabled = true
        ORDER BY priority ASC;
`

	rows, err := p.pool.Query(ctx, query, tenantID, phase)
	if err != nil {
		return nil, fmt.Errorf("query guardrail rules: %w", err)
	}
	defer rows.Close()

	out := make([]Rule, 0)

	for rows.Next() {
		var rl Rule

		if err := rows.Scan(
			&rl.ID,
			&rl.TenantID,
			&rl.Phase,
			&rl.Kind,
			&rl.Action,
			&rl.Pattern,
			&rl.Replacement,
			&rl.Priority,
			&rl.Enabled,
		); err != nil {
			return nil, fmt.Errorf("scan guardrail rule: %w", err)
		}

		out = append(out, rl)
	}

	return out, nil
}
