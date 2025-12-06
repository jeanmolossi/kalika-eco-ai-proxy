package config

import "time"

// Services declares the base URLs used by services to call each other when
// running in a distributed setup.
type Services struct {
	GatewayURL string `env:"GATEWAY_URL"       envDefault:"https://localhost:8080/api"`
	TenantURL  string `env:"TENANT_URL"        envDefault:"https://localhost:8082/api"`
	GuardURL   string `env:"GUARDRAIL_URL"     envDefault:"https://localhost:8083/api"`
	ObsURL     string `env:"OBSERVABILITY_URL" envDefault:"https://localhost:8084/api"`
	CACertFile string `env:"CA_CERTFILE"       envDefault:"cert.pem"`

	AuthToken       string        `env:"AUTH_TOKEN"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT"  envDefault:"10s"`
	MaxRetries      int           `env:"MAX_RETRIES"      envDefault:"3"`
	CircuitFailures uint32        `env:"CIRCUIT_FAILURES" envDefault:"5"`
	CircuitReset    time.Duration `env:"CIRCUIT_RESET"    envDefault:"30s"`
}
