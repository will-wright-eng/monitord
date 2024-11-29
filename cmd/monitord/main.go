package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/will-wright-eng/monitord/internal/app"
	"github.com/will-wright-eng/monitord/internal/config"
)

func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "", log.LstdFlags)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	// Log the initial configuration
	logger.Printf("Initial configuration loaded:\n")
	logger.Printf("  Database Path: %s\n", cfg.Database.Path)
	logger.Printf("  Number of Endpoints: %d\n", len(cfg.Monitor.Endpoints))
	logger.Printf("  Config Check Interval: %v\n", cfg.Monitor.ConfigCheck.ToDuration())
	logger.Printf("  Log Path: %s\n", cfg.Logging.Path)

	// Create application instance
	application, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize application: %v", err)
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start the application
	if err := application.Start(ctx); err != nil {
		logger.Fatalf("Failed to start application: %v", err)
	}

	// Wait for shutdown signal
	<-sigChan
	logger.Println("Shutdown signal received")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Graceful shutdown
	if err := application.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Failed to shutdown gracefully: %v", err)
	}
}
