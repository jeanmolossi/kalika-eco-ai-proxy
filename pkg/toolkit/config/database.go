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

// WithDefaultsFrom backfills empty configuration values using the provided base config.
// The DSN field is intentionally not inherited to avoid pointing multiple modules to
// the same database unintentionally. Instead, shared host/user/password defaults are
// applied when the module-specific configuration omits them.
func (cfg Postgres) WithDefaultsFrom(base Postgres) Postgres {
	merged := cfg

	if merged.Host == "" {
		merged.Host = base.Host
	}

	if merged.Port == 0 {
		merged.Port = base.Port
	}

	if merged.User == "" {
		merged.User = base.User
	}

	if merged.Pass == "" {
		merged.Pass = base.Pass
	}

	if merged.Database.Database == "" {
		merged.Database.Database = base.Database.Database
	}

	if merged.SSLMode == "" {
		merged.SSLMode = base.SSLMode
	}

	if merged.MaxConns == 0 {
		merged.MaxConns = base.MaxConns
	}

	if merged.MinConns == 0 {
		merged.MinConns = base.MinConns
	}

	if merged.MaxConnLifetime == 0 {
		merged.MaxConnLifetime = base.MaxConnLifetime
	}

	if merged.MaxConnIdletime == 0 {
		merged.MaxConnIdletime = base.MaxConnIdletime
	}

	if merged.HealthcheckFreq == 0 {
		merged.HealthcheckFreq = base.HealthcheckFreq
	}

	if merged.ConnectTimeout == 0 {
		merged.ConnectTimeout = base.ConnectTimeout
	}

	if merged.QueryTimeout == 0 {
		merged.QueryTimeout = base.QueryTimeout
	}

	if merged.AppName == "" {
		merged.AppName = base.AppName
	}

	return merged
}
