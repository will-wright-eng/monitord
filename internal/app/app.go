package app

import (
    "context"
    "fmt"
    "log"
    "sync"

    "github.com/will-wright-eng/monitord/internal/config"
    "github.com/will-wright-eng/monitord/internal/monitor"
    "github.com/will-wright-eng/monitord/internal/storage"
)

// App represents the main application
type App struct {
    cfg     *config.Config
    monitor *monitor.Service
    storage storage.Storage
    logger  *log.Logger
    wg      sync.WaitGroup
}

// New creates a new application instance
func New(cfg *config.Config, logger *log.Logger) (*App, error) {
    store, err := storage.NewSQLiteStore(cfg.Database.Path)
    if err != nil {
        return nil, err
    }

    // Create reload function
    reloadFn := func() (*config.Config, error) {
        return config.Load()
    }

    monitorService := monitor.NewService(
        store,
        logger,
        cfg.Monitor,
        reloadFn,
    )

    return &App{
        cfg:     cfg,
        monitor: monitorService,
        storage: store,
        logger:  logger,
    }, nil
}

// Start initializes and starts all application components
func (a *App) Start(ctx context.Context) error {
    a.logger.Println("Starting application...")

    if err := a.monitor.Start(ctx); err != nil {
        return fmt.Errorf("failed to start monitor service: %w", err)
    }

    return nil
}

// Shutdown gracefully stops all application components
func (a *App) Shutdown(ctx context.Context) error {
    a.logger.Println("Shutting down application...")

    if err := a.monitor.Shutdown(ctx); err != nil {
        a.logger.Printf("Error shutting down monitor service: %v", err)
    }

    return a.storage.Close()
}
