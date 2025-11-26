package runtime

import (
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/gateway"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
)

// Registry wires the modules required to run the gateway executable.
func Registry() core.Registry {
	return core.NewRegistry(
		database.NewModule(),
		platform.NewModule(),
		gateway.NewModule(),
	)
}

// HTTPServerConfig builds the HTTP server configuration from the shared config.
func HTTPServerConfig(cfg *toolkitconfig.Config) httpx.Config {
	return httpx.Config{
		Host:                cfg.Server.Host,
		Port:                cfg.Server.Port,
		EnableTLS:           cfg.Server.EnableTLS,
		TLSCertFile:         cfg.Server.TLSCertFile,
		TLSKeyFile:          cfg.Server.TLSKeyFile,
		BasePath:            cfg.Server.BasePath,
		ReadTimeout:         cfg.Server.ReadTimeout,
		AllowedOrigins:      cfg.Server.AllowedOrigins,
		ReadHeaderTimeout:   2 * time.Second,
		WriteTimeout:        10 * time.Second,
		IdleTimeout:         30 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		MaxRequestBodyBytes: 1 << 20, // 1MiB
	}
}
