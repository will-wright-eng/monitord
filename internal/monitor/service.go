package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/will-wright-eng/monitord/internal/config"
)

// NewService creates a new monitor service
func NewService(storage Storage, logger *log.Logger, cfg config.MonitorConfig, reloadFn func() (*config.Config, error)) *Service {
	return &Service{
		storage:   storage,
		logger:    logger,
		config:    cfg,
		endpoints: make(map[string]*EndpointMonitor),
		onReload:  reloadFn,
	}
}

// Start begins monitoring all configured endpoints
func (s *Service) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Start monitoring each enabled endpoint
	for _, endpoint := range s.config.Endpoints {
		if !endpoint.Enabled {
			continue
		}

		if err := s.startEndpoint(ctx, endpoint); err != nil {
			return fmt.Errorf("failed to start endpoint %s: %w", endpoint.URL, err)
		}
	}

	// Start configuration watcher
	s.shutdownWg.Add(1)
	go s.watchConfig(ctx)

	return nil
}

// startEndpoint begins monitoring a single endpoint
func (s *Service) startEndpoint(ctx context.Context, endpoint config.Endpoint) error {
	endpointCtx, cancel := context.WithCancel(ctx)
	monitor := &EndpointMonitor{
		endpoint: endpoint,
		cancel:   cancel,
	}
	s.endpoints[endpoint.URL] = monitor

	s.shutdownWg.Add(1)
	go s.monitorEndpoint(endpointCtx, monitor)

	return nil
}

// monitorEndpoint performs the actual health checks for an endpoint
func (s *Service) monitorEndpoint(ctx context.Context, monitor *EndpointMonitor) {
	defer s.shutdownWg.Done()

	ticker := time.NewTicker(monitor.endpoint.Interval.ToDuration())
	defer ticker.Stop()

	client := &http.Client{
		Timeout: monitor.endpoint.Timeout.ToDuration(),
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("Stopping monitoring for endpoint: %s", monitor.endpoint.URL)
			return
		case <-ticker.C:
			check := s.performHealthCheck(client, monitor.endpoint)
			if err := s.storage.SaveCheck(check); err != nil {
				s.logger.Printf("Error saving check for %s: %v", monitor.endpoint.URL, err)
			}
			monitor.lastCheck = check.Timestamp
		}
	}
}

// performHealthCheck executes a single health check
func (s *Service) performHealthCheck(client *http.Client, endpoint config.Endpoint) HealthCheck {
	s.logger.Printf("Starting health check for endpoint: %s", endpoint.URL)
	start := time.Now()
	check := HealthCheck{
		Name:      endpoint.Name,
		URL:       endpoint.URL,

		Tags:      endpoint.Tags,
		Timestamp: start,
	}

	resp, err := client.Get(endpoint.URL)
	if err != nil {
		s.logger.Printf("Error checking endpoint %s: %v", endpoint.URL, err)
		check.Status = "ERROR"
		check.Error = err.Error()
		return check
	}
	defer resp.Body.Close()

	duration := time.Since(start).Milliseconds()
	check.StatusCode = resp.StatusCode
	check.ResponseTime = duration

	if resp.StatusCode == http.StatusOK {
		check.Status = "UP"
		s.logger.Printf("Health check successful for %s - Status: %s, Response time: %dms",
			endpoint.URL, check.Status, duration)
	} else {
		check.Status = "DEGRADED"
		s.logger.Printf("Health check degraded for %s - Status: %s, Status code: %d, Response time: %dms",
			endpoint.URL, check.Status, resp.StatusCode, duration)
	}

	return check
}

// watchConfig periodically checks for configuration updates
func (s *Service) watchConfig(ctx context.Context) {
	defer s.shutdownWg.Done()

	ticker := time.NewTicker(s.config.ConfigCheck.ToDuration())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.reloadConfig(); err != nil {
				s.logger.Printf("Error reloading configuration: %v", err)
			}
		}
	}
}

// reloadConfig reloads the configuration
func (s *Service) reloadConfig() error {
	s.logger.Println("Reloading configuration...")

	cfg, err := s.onReload()
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Track new endpoints
	newEndpoints := make(map[string]*EndpointMonitor)

	// Process new or existing endpoints
	for _, endpoint := range cfg.Monitor.Endpoints {
		if !endpoint.Enabled {
			continue
		}

		// Check if endpoint already exists
		if monitor, exists := s.endpoints[endpoint.URL]; exists {
			// Update existing endpoint if configuration changed
			if !endpointConfigEqual(monitor.endpoint, endpoint) {
				s.logger.Printf("Updating configuration for endpoint: %s", endpoint.URL)
				monitor.cancel()
				if err := s.startEndpoint(context.Background(), endpoint); err != nil {
					return fmt.Errorf("failed to restart endpoint %s: %w", endpoint.URL, err)
				}
			}
			newEndpoints[endpoint.URL] = s.endpoints[endpoint.URL]
		} else {
			// Start monitoring new endpoint
			s.logger.Printf("Adding new endpoint: %s", endpoint.URL)
			if err := s.startEndpoint(context.Background(), endpoint); err != nil {
				return fmt.Errorf("failed to start new endpoint %s: %w", endpoint.URL, err)
			}
			newEndpoints[endpoint.URL] = s.endpoints[endpoint.URL]
		}
	}

	// Stop monitoring removed endpoints
	for url, monitor := range s.endpoints {
		if _, exists := newEndpoints[url]; !exists {
			s.logger.Printf("Removing endpoint: %s", url)
			monitor.cancel()
		}
	}

	// Update endpoints map
	s.endpoints = newEndpoints
	s.config = cfg.Monitor

	return nil
}

// Shutdown gracefully stops all monitoring
func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	// Cancel all endpoint monitors
	for _, monitor := range s.endpoints {
		monitor.cancel()
	}
	s.mu.Unlock()

	// Wait for all goroutines to finish or context to cancel
	done := make(chan struct{})
	go func() {
		s.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// endpointConfigEqual compares two endpoint configurations
func endpointConfigEqual(a, b config.Endpoint) bool {
	return a.URL == b.URL &&
		a.Interval == b.Interval &&
		a.Timeout == b.Timeout &&
		a.Name == b.Name &&
		sliceEqual(a.Tags, b.Tags)
}

// sliceEqual compares two string slices
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
