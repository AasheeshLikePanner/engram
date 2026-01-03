package main

import (
	"engram/internal/agent"
	"engram/internal/config"
	"engram/internal/server"
	"engram/internal/store"
	pb "engram/proto/v1"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Load()

	var s store.DurableStore
	var err error

	if cfg.StoreType == "redis" {
		log.Printf("Initializing Redis store at %s", cfg.RedisURL)
		s, err = store.NewRedisStore(cfg.RedisURL)
	} else {
		log.Printf("Initializing Badger store at %s", cfg.DBPath)
		s, err = store.NewBadgerStore(cfg.DBPath)
	}

	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer s.Close()

	// Initialize Agent Service
	agentSvc := agent.NewAgent(s)

	// Initialize Server
	srv, err := server.NewServer(cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// Register Service
	pb.RegisterAgentServiceServer(srv.GrpcServer, agentSvc)

	// Graceful Shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		srv.Stop()
	}()

	log.Printf("Starting gRPC server on :%d", cfg.GRPCPort)
	if err := srv.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
