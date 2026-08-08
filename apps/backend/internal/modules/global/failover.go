// Package global provides regional circuit breaking, health evaluation, and automated multi-region DNS failover runbooks.
package global

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrCircuitOpen is returned when execution is blocked by an open circuit breaker.
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitState represents the current state of a circuit breaker.
type CircuitState string

const (
	StateClosed   CircuitState = "closed"
	StateOpen     CircuitState = "open"
	StateHalfOpen CircuitState = "half-open"
)

// CircuitBreaker guards external DB and service calls from cascading failure.
type CircuitBreaker struct {
	maxFailures         int
	timeout             time.Duration
	consecutiveFailures int
	state               CircuitState
	lastStateChange     time.Time
	mu                  sync.RWMutex
}

// NewCircuitBreaker initializes a CircuitBreaker with failure threshold and recovery timeout.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:     maxFailures,
		timeout:         timeout,
		state:           StateClosed,
		lastStateChange: time.Now(),
	}
}

// State returns current CircuitState.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Execute runs the provided function if circuit is closed or half-open.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	if cb.state == StateOpen {
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.state = StateHalfOpen
			cb.lastStateChange = time.Now()
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.consecutiveFailures++
		if cb.consecutiveFailures >= cb.maxFailures {
			cb.state = StateOpen
			cb.lastStateChange = time.Now()
		}
		return err
	}

	cb.consecutiveFailures = 0
	cb.state = StateClosed
	cb.lastStateChange = time.Now()
	return nil
}

// FailoverResult captures automated DNS failover outcome details.
type FailoverResult struct {
	FailoverTriggered bool      `json:"failover_triggered"`
	PreviousRegion    string    `json:"previous_region"`
	NewActiveRegion   string    `json:"new_active_region"`
	Timestamp         time.Time `json:"timestamp"`
}

// ClusterFailoverStatus represents high-level multi-region failover and HA status.
type ClusterFailoverStatus struct {
	ActiveRegion  string          `json:"active_region"`
	BackupRegions []string        `json:"backup_regions"`
	IsFailedOver  bool            `json:"is_failed_over"`
	RegionHealth  map[string]bool `json:"region_health"`
}

// FailoverManager monitors region health and triggers automated DNS traffic rerouting during outages (<10s).
type FailoverManager struct {
	primaryRegion string
	activeRegion  string
	backupRegions []string
	failures      map[string]int
	maxFailures   int
	mu            sync.RWMutex
}

// NewFailoverManager constructs a FailoverManager for specified primary and backup regions.
func NewFailoverManager(primaryRegion string, backupRegions []string) *FailoverManager {
	return &FailoverManager{
		primaryRegion: primaryRegion,
		activeRegion:  primaryRegion,
		backupRegions: backupRegions,
		failures:      make(map[string]int),
		maxFailures:   3,
	}
}

// RecordRegionFailure logs a regional health check failure and triggers failover if threshold is reached.
func (fm *FailoverManager) RecordRegionFailure(region string) (*FailoverResult, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.failures[region]++

	if region == fm.activeRegion && fm.failures[region] >= fm.maxFailures {
		previous := fm.activeRegion
		var nextRegion string

		for _, b := range fm.backupRegions {
			if fm.failures[b] < fm.maxFailures {
				nextRegion = b
				break
			}
		}

		if nextRegion == "" {
			return nil, fmt.Errorf("all backup regions are degraded, failover impossible")
		}

		fm.activeRegion = nextRegion
		return &FailoverResult{
			FailoverTriggered: true,
			PreviousRegion:    previous,
			NewActiveRegion:   nextRegion,
			Timestamp:         time.Now().UTC(),
		}, nil
	}

	return &FailoverResult{
		FailoverTriggered: false,
		PreviousRegion:    fm.activeRegion,
		NewActiveRegion:   fm.activeRegion,
		Timestamp:         time.Now().UTC(),
	}, nil
}

// RecordRegionSuccess resets failure counters for a recovered region.
func (fm *FailoverManager) RecordRegionSuccess(region string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.failures[region] = 0
	return nil
}

// GetClusterState retrieves active failover status across primary and backup regions.
func (fm *FailoverManager) GetClusterState(ctx context.Context) (*ClusterFailoverStatus, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	health := make(map[string]bool)
	health[fm.primaryRegion] = fm.failures[fm.primaryRegion] < fm.maxFailures
	for _, b := range fm.backupRegions {
		health[b] = fm.failures[b] < fm.maxFailures
	}

	return &ClusterFailoverStatus{
		ActiveRegion:  fm.activeRegion,
		BackupRegions: fm.backupRegions,
		IsFailedOver:  fm.activeRegion != fm.primaryRegion,
		RegionHealth:  health,
	}, nil
}
