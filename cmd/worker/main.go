package main

import (
	"context"
	"engram/internal/config"
	"engram/internal/llm"
	"engram/internal/store"
	"engram/internal/worker"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Worker failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()

	var db store.DurableStore
	var err error

	if cfg.StoreType == "redis" {
		slog.Info("Initializing Redis store", "url", cfg.RedisURL)
		db, err = store.NewRedisStore(cfg.RedisURL)
	} else {
		slog.Info("Initializing Badger store", "path", cfg.DBPath)
		db, err = store.NewBadgerStore(cfg.DBPath)
	}

	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}
	defer db.Close()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "worker-1"
	}

	var llmClient llm.LLM
	if cfg.LLMProvider == "openai" {
		slog.Info("Using OpenAI-compatible LLM provider", "url", cfg.LLMBaseURL)
		llmClient = llm.NewOpenAI(cfg.LLMAPIKey, cfg.LLMBaseURL)
	} else {
		slog.Info("Using Ollama LLM provider", "url", cfg.OllamaURL)
		llmClient = llm.NewOllama(cfg.OllamaURL)
	}
	
	w := worker.NewWorker(db, llmClient, hostname, cfg.ToolsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)
	go func() {
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			errChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		slog.Info("Received signal, shutting down", "signal", sig)
		cancel()
	case err := <-errChan:
		return fmt.Errorf("worker error: %w", err)
	}

	return nil
}
