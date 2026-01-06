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
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Combined engine failed", "error", err)
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

	// 2. Manager
	agentSvc := agent.NewAgent(s)
	srv, err := server.NewServer(cfg.GRPCPort, cfg.APIKey)
	if err != nil {
		return fmt.Errorf("failed to initialize server: %w", err)
	}
	pb.RegisterAgentServiceServer(srv.GrpcServer, agentSvc)

	// 3. Worker
	var llmClient llm.LLM
	if cfg.LLMProvider == "openai" {
		slog.Info("Using OpenAI-compatible LLM provider", "url", cfg.LLMBaseURL)
		llmClient = llm.NewOpenAI(cfg.LLMAPIKey, cfg.LLMBaseURL)
	} else {
		slog.Info("Using Ollama LLM provider", "url", cfg.OllamaURL)
		llmClient = llm.NewOllama(cfg.OllamaURL)
	}

	w := worker.NewWorker(s, llmClient, "combined-worker", cfg.ToolsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start everything
	errChan := make(chan error, 2)
	go func() {
		slog.Info("Starting gRPC Manager", "port", cfg.GRPCPort)
		if err := srv.Serve(); err != nil {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	go func() {
		slog.Info("Starting Worker")
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			errChan <- fmt.Errorf("worker error: %w", err)
		}
	}()

	// Graceful Shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
		srv.Stop()
	case err := <-errChan:
		return err
	}

	return nil
}
