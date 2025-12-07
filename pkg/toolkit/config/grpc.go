package config

import "time"

type GRPCServer struct {
	Host             string        `env:"HOST"               envDefault:"0.0.0.0"`
	Port             int           `env:"PORT"               envDefault:"9091"`
	EnableTLS        bool          `env:"ENABLE_TLS"         envDefault:"false"`
	TLSCertFile      string        `env:"TLS_CERTFILE"       envDefault:""`
	TLSKeyFile       string        `env:"TLS_KEYFILE"        envDefault:""`
	MaxRecvMsgBytes  int           `env:"MAX_RECV_MSG_BYTES" envDefault:"4194304"`
	MaxSendMsgBytes  int           `env:"MAX_SEND_MSG_BYTES" envDefault:"4194304"`
	EnableReflection bool          `env:"ENABLE_REFLECTION"  envDefault:"true"`
	Enabled          bool          `env:"ENABLED"            envDefault:"true"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT"   envDefault:"15s"`
}

func (g *GRPCServer) Normalize() {
	if g.Port == 0 {
		g.Port = 9091
	}

	if g.MaxRecvMsgBytes <= 0 {
		g.MaxRecvMsgBytes = 4 << 20
	}

	if g.MaxSendMsgBytes <= 0 {
		g.MaxSendMsgBytes = 4 << 20
	}

	if g.ShutdownTimeout <= 0 {
		g.ShutdownTimeout = 15 * time.Second
	}
}
