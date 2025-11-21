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

	Log       Log        `envPrefix:"LOG_"`
	Server    HTTPServer `envPrefix:"SERVER_"`
	PgDB      Postgres   `envPrefix:"POSTGRES_"`
	RateLimit RateLimit  `envPrefix:"RATELIMIT_"`
}

var Load = sync.OnceValue(loadEnv)

func loadEnv() *Config {
	var c Config

	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	return &c
}

func ResetForTests() {
	Load = sync.OnceValue(loadEnv)
}
