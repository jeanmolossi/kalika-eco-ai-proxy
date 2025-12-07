package config

import (
	"strings"
	"time"
)

type HTTPServer struct {
	Host                string        `env:"HOST"                   envDefault:"0.0.0.0"`
	Port                int           `env:"PORT"                   envDefault:"8081"`
	BasePath            string        `env:"BASE_PATH"              envDefault:"/api"`
	ReadTimeout         time.Duration `env:"READ_TIMEOUT"           envDefault:"5s"`
	AllowedOrigins      []string      `env:"ALLOWED_ORIGINS"        envSeparator:","`
	MaxRequestBodyBytes int64         `env:"MAX_REQUEST_BODY_BYTES" envDefault:"20971520"`

	TLSCertFile string `env:"TLS_CERTFILE" envDefault:"cert.pem"`
	TLSKeyFile  string `env:"TLS_KEYFILE"  envDefault:"key.pem"`
	EnableTLS   bool   `env:"ENABLE_TLS"   envDefault:"true"`
}

// Normalize ensures the configured base path is standardized before being used across the app.
// It guarantees a single leading slash and removes trailing slashes, except when the base path
// represents the root.
func (h *HTTPServer) Normalize() {
	h.BasePath = NormalizeBasePath(h.BasePath)

	if h.MaxRequestBodyBytes <= 0 {
		h.MaxRequestBodyBytes = 20 << 20 // 20MiB
	}
}

// NormalizeBasePath standardizes base paths used by the HTTP server.
func NormalizeBasePath(basePath string) string {
	basePath = strings.TrimSpace(basePath)

	if basePath == "" || basePath == "/" {
		return ""
	}

	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}

	return strings.TrimSuffix(basePath, "/")
}
