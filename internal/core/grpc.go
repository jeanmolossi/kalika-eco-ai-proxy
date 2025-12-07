package core

import "google.golang.org/grpc"

// GRPCModule allows modules to register gRPC services when available.
type GRPCModule interface {
	RegisterGRPC(server *grpc.Server, c *Container) error
}
