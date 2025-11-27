package runtime

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/cache"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/gateway"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/llm"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/ratelimit"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant"
)

// Registry wires the modules required to run the gateway executable.
func Registry() core.Registry {
	return core.NewRegistry(
		database.NewModule(),
		tenant.NewModule(),
		guardrails.NewModule(),
		ratelimit.NewModule(),
		cache.NewModule(),
		llm.NewModule(),
		observability.NewModule(),
		gateway.NewModule(),
	)
}
