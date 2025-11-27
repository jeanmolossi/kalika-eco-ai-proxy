package httpx

import (
	"time"

	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
)

// FromHTTPServer builds the HTTP server settings from a specific HTTPServer
// configuration so individual services can bootstrap with consistent defaults.
func FromHTTPServer(server toolkitconfig.HTTPServer) Config {
	return Config{
		Host:                server.Host,
		Port:                server.Port,
		EnableTLS:           server.EnableTLS,
		TLSCertFile:         server.TLSCertFile,
		TLSKeyFile:          server.TLSKeyFile,
		BasePath:            server.BasePath,
		ReadTimeout:         server.ReadTimeout,
		AllowedOrigins:      server.AllowedOrigins,
		ReadHeaderTimeout:   2 * time.Second,
		WriteTimeout:        10 * time.Second,
		IdleTimeout:         30 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		MaxRequestBodyBytes: 1 << 20, // 1MiB
	}
}
