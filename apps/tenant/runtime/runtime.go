package runtime

import (
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/database"
	tenantmodule "github.com/jeanmolossi/kalika-eco-ai-proxy/internal/tenant/module"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
)

// Registry wires the modules required to run the tenant service.
func Registry() core.Registry {
	return core.NewRegistry(
		database.NewTenantModule(),
		tenantmodule.NewModule(),
	)
}

// HTTPServerConfig builds the HTTP server configuration for the tenant API.
func HTTPServerConfig(cfg *toolkitconfig.Config) httpx.Config {
	return httpx.FromHTTPServer(cfg.Tenant)
}
