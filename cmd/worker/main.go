package main

import (
	"context"
	"engram/internal/config"
	"engram/internal/llm"
	"engram/internal/store"
	"engram/internal/worker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.Load()

	var db store.DurableStore
	var err error

	if cfg.StoreType == "redis" {
		log.Printf("Initializing Redis store at %s", cfg.RedisURL)
		db, err = store.NewRedisStore(cfg.RedisURL)
	} else {
		log.Printf("Initializing Badger store at %s", cfg.DBPath)
		db, err = store.NewBadgerStore(cfg.DBPath)
	}

	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer db.Close()

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "worker-1"
	}

	// llmClient := llm.NewOllama(cfg.OllamaURL)
	llmClient := &llm.MockLLM{}
	w := worker.NewWorker(db, llmClient, hostname, cfg.ToolsPath)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := w.Start(ctx); err != nil && err != context.Canceled {
			log.Printf("Worker error: %v", err)
		}
	}()

	<-sigChan
}
