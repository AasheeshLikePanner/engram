package server

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type Server struct {
	GrpcServer *grpc.Server
	Listener   net.Listener
}

func NewServer(port int) (*Server, error) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	grpcServer := grpc.NewServer()
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
