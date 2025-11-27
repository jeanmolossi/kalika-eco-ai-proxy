package runtime

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/guardrails"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
)

// Registry wires the modules required to run the guardrails service.
func Registry() core.Registry {
	return core.NewRegistry(
		database.NewModule(),
		guardrails.NewModule(),
	)
}

// HTTPServerConfig builds the HTTP server configuration for the guardrails API.
func HTTPServerConfig(cfg *toolkitconfig.Config) httpx.Config {
	return httpx.FromToolkitConfig(cfg)
}
