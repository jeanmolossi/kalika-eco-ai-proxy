package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Root configs

	Environment       Environment `env:"ENVIRONMENT"        envDefault:"production"`
	GuardrailsEnabled bool        `env:"GUARDRAILS_ENABLED" envDefault:"true"`

	// Nested configs

	Log     Log        `envPrefix:"LOG_"`
	Server  HTTPServer `envPrefix:"SERVER_"`
	Gateway HTTPServer `envPrefix:"GATEWAY_SERVER_"`
	Tenant  HTTPServer `envPrefix:"TENANT_SERVER_"`
	Guard   HTTPServer `envPrefix:"GUARDRAIL_SERVER_"`
	Observe HTTPServer `envPrefix:"OBSERVABILITY_SERVER_"`

	LLM       LLM       `envPrefix:"LLM_"`
	PgDB      Postgres  `envPrefix:"POSTGRES_"`
	RateLimit RateLimit `envPrefix:"RATELIMIT_"`
	Kafka     Kafka     `envPrefix:"KAFKA_"`
	UsageSink UsageSink `envPrefix:"USAGE_"`
	AuditSink AuditSink `envPrefix:"AUDIT_"`

	Services Services `envPrefix:"SERVICES_"`
}

var Load = sync.OnceValue(loadEnv)

func loadEnv() *Config {
	var c Config

	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	c.Server.Normalize()
	c.Gateway.Normalize()
	c.Tenant.Normalize()
	c.Guard.Normalize()
	c.Observe.Normalize()

	return &c
}

func ResetForTests() {
	Load = sync.OnceValue(loadEnv)
}
