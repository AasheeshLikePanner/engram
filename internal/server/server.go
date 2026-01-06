package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	GrpcServer *grpc.Server
	Listener   net.Listener
}

func NewServer(port int, apiKey string) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	var opts []grpc.ServerOption

	if apiKey != "" {
		authInterceptor := func(ctx context.Context) error {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return status.Error(codes.Unauthenticated, "metadata is missing")
			}
			keys := md.Get("x-api-key")
			if len(keys) == 0 || keys[0] != apiKey {
				return status.Error(codes.Unauthenticated, "invalid api key")
			}
			return nil
		}

		opts = append(opts, grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			if err := authInterceptor(ctx); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}))

		opts = append(opts, grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if err := authInterceptor(ss.Context()); err != nil {
				return err
			}
			return handler(srv, ss)
		}))
	}

	grpcServer := grpc.NewServer(opts...)
	return &Server{
		GrpcServer: grpcServer,
		Listener:   lis,
	}, nil
}

func (s *Server) Serve() error {
	return s.GrpcServer.Serve(s.Listener)
}

func (s *Server) Stop() {
	s.GrpcServer.GracefulStop()
}
