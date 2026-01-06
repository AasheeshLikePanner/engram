package main

import (
	"engram/internal/agent"
	"engram/internal/config"
	"engram/internal/server"
	"engram/internal/store"
	pb "engram/proto/v1"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	var s store.DurableStore
	var err error

	if cfg.StoreType == "redis" {
		slog.Info("Initializing Redis store", "url", cfg.RedisURL)
		s, err = store.NewRedisStore(cfg.RedisURL)
	} else {
		slog.Info("Initializing Badger store", "path", cfg.DBPath)
		s, err = store.NewBadgerStore(cfg.DBPath)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer s.Close()

	// Initialize Agent Service
	agentSvc := agent.NewAgent(s)

	// Initialize Server
	srv, err := server.NewServer(cfg.GRPCPort, cfg.APIKey)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}

	// Register Service
	pb.RegisterAgentServiceServer(srv.GrpcServer, agentSvc)

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		slog.Info("Starting gRPC server", "port", cfg.GRPCPort)
		if err := srv.Serve(); err != nil {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", sig)
		srv.Stop()
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
