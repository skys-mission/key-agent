// Package main is the entry point for key-agent daemon.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/skys-mission/key-agent/internal/config"
	"github.com/skys-mission/key-agent/internal/daemon"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.ShowVersion {
		fmt.Printf("key-agent %s (commit: %s, built: %s)\n", version, commit, date)
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, gracefully stopping...")
		cancel()
	}()

	if err := daemon.Run(ctx, cfg); err != nil {
		log.Fatalf("Daemon error: %v", err)
	}
}
