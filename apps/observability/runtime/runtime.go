package runtime

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/observability"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
)

// Registry wires the modules required to run the observability service.
func Registry() core.Registry {
	return core.NewRegistry(
		database.NewModule(),
		observability.NewModule(),
	)
}

// HTTPServerConfig builds the HTTP server configuration for the observability API.
func HTTPServerConfig(cfg *toolkitconfig.Config) httpx.Config {
	return httpx.FromHTTPServer(cfg.Observe)
}
