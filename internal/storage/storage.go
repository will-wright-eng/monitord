package storage

import (
	"github.com/will-wright-eng/monitord/internal/monitor"
)

// Storage defines the interface for storing health check results
type Storage interface {
	SaveCheck(check monitor.HealthCheck) error
	Close() error
}