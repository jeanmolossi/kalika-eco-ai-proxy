package httpx

import (
	"time"

	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
)

// FromToolkitConfig builds the HTTP server settings from the shared toolkit
// configuration so individual services can bootstrap with consistent defaults.
func FromToolkitConfig(cfg *toolkitconfig.Config) Config {
	return Config{
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
