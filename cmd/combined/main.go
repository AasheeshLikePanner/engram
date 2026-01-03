package main

import (
	"context"
	"engram/internal/agent"
	"engram/internal/config"
	"engram/internal/llm"
	"engram/internal/server"
	"engram/internal/store"
	"engram/internal/worker"
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

	// 2. Manager
	agentSvc := agent.NewAgent(s)
	srv, err := server.NewServer(cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}
	pb.RegisterAgentServiceServer(srv.GrpcServer, agentSvc)

	// 3. Worker (Using MockLLM for verification)
	llmClient := &llm.MockLLM{}
	w := worker.NewWorker(s, llmClient, "combined-worker")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start everything
	go func() {
		log.Printf("Starting gRPC Manager on :%d", cfg.GRPCPort)
		if err := srv.Serve(); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	go func() {
		log.Println("Starting reactive Worker...")
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Worker error: %v", err)
		}
	}()

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down...")
	srv.Stop()
}
