package monitor

import (
    "sync"
    "time"
)

type ShutdownManager struct {
    wg      sync.WaitGroup
    timeout time.Duration
}

func NewShutdownManager() *ShutdownManager {
    return &ShutdownManager{
        timeout: 5 * time.Second,
    }
}

func (sm *ShutdownManager) Add(delta int) {
    sm.wg.Add(delta)
}

func (sm *ShutdownManager) Done() {
    sm.wg.Done()
}

func (sm *ShutdownManager) Shutdown() bool {
    done := make(chan struct{})
    go func() {
        sm.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return true
    case <-time.After(sm.timeout):
        return false
	}
}
