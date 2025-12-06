package guardrails

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database/pg"
)

func ProvideGuardrails(c *core.Container) Engine {
	conf := c.Config()
	logger := c.Logger().With("module", "guardrails")

	if !conf.GuardrailsEnabled {
		return NewNoopGuardrails()
	}

	db := core.MustGet[*pg.DB](c, database.GuardrailConn)

	repo := NewPGRuleRepository(db.Pool())
	sink := NewLoggerSink(logger)

	return NewSimpleEngine(repo, logger, sink)
}
