package grpcx

import (
	"time"

	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
)

type Config struct {
	Host             string
	Port             int
	EnableTLS        bool
	TLSCertFile      string
	TLSKeyFile       string
	MaxRecvMsgBytes  int
	MaxSendMsgBytes  int
	EnableReflection bool
	Enabled          bool
	ShutdownTimeout  time.Duration
}

func FromGRPCServer(server toolkitconfig.GRPCServer) Config {
	return Config{
		Host:             server.Host,
		Port:             server.Port,
		EnableTLS:        server.EnableTLS,
		TLSCertFile:      server.TLSCertFile,
		TLSKeyFile:       server.TLSKeyFile,
		MaxRecvMsgBytes:  server.MaxRecvMsgBytes,
		MaxSendMsgBytes:  server.MaxSendMsgBytes,
		EnableReflection: server.EnableReflection,
		Enabled:          server.Enabled,
		ShutdownTimeout:  server.ShutdownTimeout,
	}
}
