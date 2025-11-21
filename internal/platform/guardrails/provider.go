package guardrails

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database/pg"
)

func ProvideGuardrails(c *core.Container) Engine {
	conf := c.Config()
	logger := c.Logger().With("module", "guardrails")

	if !conf.GuardrailsEnabled {
		return NewNoopGuardrails()
	}

	db := core.MustGet[*pg.DB](c, database.PgConn)

	repo := NewPGRuleRepository(db.Pool())

	return NewSimpleEngine(repo, logger)
}
