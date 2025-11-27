package config

// Services declares the base URLs used by services to call each other when
// running in a distributed setup.
type Services struct {
	GatewayURL string `env:"GATEWAY_URL" envDefault:"http://localhost:8080/api"`
	TenantURL  string `env:"TENANT_URL"  envDefault:"http://localhost:8082/api"`
	GuardURL   string `env:"GUARDRAIL_URL" envDefault:"http://localhost:8083/api"`
	ObsURL     string `env:"OBSERVABILITY_URL" envDefault:"http://localhost:8084/api"`
}
