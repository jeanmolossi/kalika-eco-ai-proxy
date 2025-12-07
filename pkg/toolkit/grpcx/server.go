package grpcx

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewServer(cfg Config) (*grpc.Server, *health.Server, error) {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(cfg.MaxRecvMsgBytes),
		grpc.MaxSendMsgSize(cfg.MaxSendMsgBytes),
	}

	if cfg.EnableTLS {
		creds, err := credentials.NewServerTLSFromFile(cfg.TLSCertFile, cfg.TLSKeyFile)
		if err != nil {
			return nil, nil, fmt.Errorf("load grpc tls credentials: %w", err)
		}

		opts = append(opts, grpc.Creds(creds))
	}

	server := grpc.NewServer(opts...)
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(server, healthSrv)

	if cfg.EnableReflection {
		reflection.Register(server)
	}

	return server, healthSrv, nil
}

func Start(cfg Config) func(context.Context, *grpc.Server) (func(context.Context) error, error) {
	return func(ctx context.Context, server *grpc.Server) (func(context.Context) error, error) {
		if !cfg.Enabled {
			return nil, nil
		}

		if server == nil {
			return nil, errors.New("grpc server is nil")
		}

		addr := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

		lis, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("grpc listen on %s: %w", addr, err)
		}

		errCh := make(chan error, 1)

		go func() {
			errCh <- server.Serve(lis)
		}()

		slog.Default().InfoContext(ctx, "gRPC server listening", slog.String("addr", addr))

		return func(ctx context.Context) error {
			shutdownCtx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
			defer cancel()

			stopped := make(chan struct{})

			go func() {
				server.GracefulStop()
				close(stopped)
			}()

			select {
			case err := <-errCh:
				if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
					return err
				}

				return nil
			case <-stopped:
				return nil
			case <-shutdownCtx.Done():
				server.Stop()
				return shutdownCtx.Err()
			}
		}, nil
	}
}
