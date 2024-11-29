package monitor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/will-wright-eng/monitord/internal/config"
)

// Storage interface defines the required methods for storing health checks
type Storage interface {
	SaveCheck(check HealthCheck) error
	Close() error
}

// Service handles the monitoring of endpoints
type Service struct {
	storage    Storage
	logger     *log.Logger
	config     config.MonitorConfig
	endpoints  map[string]*EndpointMonitor
	mu         sync.RWMutex
	shutdownWg sync.WaitGroup
	onReload   func() (*config.Config, error)
}

// EndpointMonitor represents an individual endpoint monitoring goroutine
type EndpointMonitor struct {
	endpoint  config.Endpoint
	cancel    context.CancelFunc
	lastCheck time.Time
}

// HealthCheck represents the result of a single health check
type HealthCheck struct {
	Name         string    `json:"name"`
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	StatusCode   int       `json:"statusCode"`
	ResponseTime int64     `json:"responseTime"`
	Timestamp    time.Time `json:"timestamp"`
	Error        string    `json:"error,omitempty"`
	Tags         []string  `json:"tags,omitempty"`
}
