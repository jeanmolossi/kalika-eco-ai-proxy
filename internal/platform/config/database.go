package config

import "time"

type Database struct {
	// Can connect via DSN

	DSN string `env:"DSN"`

	// OR connect via credentials

	Host     string `env:"HOST"     envDefault:"localhost"`
	Port     int    `env:"PORT"`
	User     string `env:"USER"`
	Pass     string `env:"PASSWORD"`
	Database string `env:"DB"`
	SSLMode  string `env:"SSL_MODE"` // disable | require | verify-full

	// Pool

	MaxConns        int32         `env:"MAX_CONNS"         envDefault:"20"`
	MinConns        int32         `env:"MIN_CONNS"         envDefault:"2"`
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME"`
	MaxConnIdletime time.Duration `env:"MAX_CONN_IDLETIME" envDefault:"30s"`
	HealthcheckFreq time.Duration `env:"HEALTHCHECK_FREQ"  envDefault:"30s"`

	// Guard timeouts

	ConnectTimeout time.Duration `env:"CONNECT_TIMEOUT" envDefault:"10s"` // handshake
	QueryTimeout   time.Duration `env:"QUERY_TIMEOUT"   envDefault:"10s"` // default to ctx with deadline if not come

	AppName string `env:"APP_NAME"` // misc
}

type Postgres struct {
	Database
}
