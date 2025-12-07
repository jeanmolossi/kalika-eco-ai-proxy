package grpcx

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfig struct {
	Address     string
	UseTLS      bool
	CACertFile  string
	ServerName  string
	DialTimeout time.Duration
}

func Dial(ctx context.Context, cfg ClientConfig) (*grpc.ClientConn, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("grpc address is required")
	}

	timeout := cfg.DialTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	opts := []grpc.DialOption{}

	if cfg.UseTLS {
		certPool := x509.NewCertPool()

		if cfg.CACertFile != "" {
			pem, err := os.ReadFile(cfg.CACertFile)
			if err != nil {
				return nil, fmt.Errorf("read ca cert: %w", err)
			}

			if ok := certPool.AppendCertsFromPEM(pem); !ok {
				return nil, fmt.Errorf("invalid ca cert file")
			}
		}

		creds := credentials.NewClientTLSFromCert(certPool, cfg.ServerName)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(dialCtx, cfg.Address, opts...) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("dial grpc %s: %w", cfg.Address, err)
	}

	return conn, nil
}
