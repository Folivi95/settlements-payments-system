package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/saltpay/settlements-payments-system/cmd/config"
)

// Pinger defines the pinger methods.
type Pinger interface {
	// Ping pings a service and returns an error if service is unhealthy
	// Implementations should handle context cancellation properly.
	Ping(ctx context.Context) error
}

// NamedPinger represents a named pinger that implements the Pinger interface.
type NamedPinger struct {
	Pinger
	Name string
}

type status struct {
	Healthy      bool            `json:"healthy"`
	Dependencies map[string]bool `json:"dependencies,omitempty"`
}

type handler struct {
	sync.RWMutex
	status        status
	checkInterval time.Duration
	checkTimeout  time.Duration
	pingers       []NamedPinger
}

// Option represents an health handler option.
type Option func(*handler)

// Pingers sets the list of pingers to check as part of the overall health.
func Pingers(pingers ...NamedPinger) Option {
	return func(h *handler) {
		h.pingers = pingers
	}
}

const (
	checkIntervalSec = 5
	checkTimeoutMs   = 500
)

// New instantiates a new health handler.
func NewHealthCheckHandler(config config.Config, opts ...Option) *handler {
	var (
		checkInterval = checkIntervalSec * time.Second
		checkTimeout  = checkTimeoutMs * time.Millisecond
	)

	if config.HealthCheckTimeout != 0 {
		checkTimeout = config.HealthCheckTimeout
	}

	if config.HealthCheckInterval != 0 {
		checkInterval = config.HealthCheckInterval
	}

	h := &handler{
		status: status{
			Healthy: false,
		},
		checkInterval: checkInterval,
		checkTimeout:  checkTimeout,
		pingers:       nil,
	}

	for _, opt := range opts {
		opt(h)
	}

	h.checkServices()

	go h.updateStatus()

	return h
}

func (h *handler) updateStatus() {
	for range time.Tick(h.checkInterval) {
		h.checkServices()
	}
}

func (h *handler) checkServices() {
	h.Lock()
	defer h.Unlock()

	type depHealth struct {
		name   string
		health bool
	}

	// TODO: should probably be the application context
	ctx, cancel := context.WithTimeout(context.Background(), h.checkTimeout)
	defer cancel()

	results := make(chan depHealth)
	done := make(chan struct{})

	if len(h.pingers) == 0 {
		h.status.Healthy = true
		return
	}

	result := status{Healthy: true}
	result.Dependencies = make(map[string]bool, len(h.pingers))
	for _, s := range h.pingers {
		result.Dependencies[s.Name] = false
	}

	go func() {
		for r := range results {
			result.Dependencies[r.name] = r.health
			if !r.health {
				result.Healthy = false
			}
		}
		done <- struct{}{}
	}()

	var wg sync.WaitGroup
	wg.Add(len(h.pingers))

	for _, s := range h.pingers {
		go func(s NamedPinger) {
			defer wg.Done()
			err := s.Ping(ctx)
			if err != nil {
				results <- depHealth{
					name:   s.Name,
					health: false,
				}
				return
			}
			results <- depHealth{
				name:   s.Name,
				health: true,
			}
		}(s)
	}

	wg.Wait()
	close(results)
	<-done

	h.status.Healthy = result.Healthy
	h.status.Dependencies = result.Dependencies
}

// HealthCheckHandler is the handler for the health endpoint. Currently, there's a goroutine
// that asynchronously performs a status check for all services every X amount of seconds.
// As a result, this handler returns a cached value of the latest execution.
func (h *handler) HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	h.RLock()
	defer h.RUnlock()

	if h.status.Healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	_, _ = fmt.Fprint(w, "Healthy")
}
