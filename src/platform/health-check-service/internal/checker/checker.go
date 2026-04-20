package checker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/shopos/health-check-service/internal/config"
	"github.com/shopos/health-check-service/internal/domain"
	"go.uber.org/zap"
)

// Checker probes all configured targets and maintains their health state.
type Checker struct {
	cfg     *config.Config
	client  *http.Client
	log     *zap.Logger
	mu      sync.RWMutex
	state   map[string]*domain.TargetHealth
	stopCh  chan struct{}
}

func New(cfg *config.Config, log *zap.Logger) *Checker {
	state := make(map[string]*domain.TargetHealth, len(cfg.Targets))
	for _, t := range cfg.Targets {
		state[t.Name] = &domain.TargetHealth{
			Name:   t.Name,
			URL:    t.URL,
			Status: domain.StatusUnknown,
		}
	}
	return &Checker{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.CheckTimeout},
		log:    log,
		state:  state,
		stopCh: make(chan struct{}),
	}
}

// Start runs periodic checks in the background until Stop is called.
func (c *Checker) Start() {
	c.runAll()
	ticker := time.NewTicker(c.cfg.CheckInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.runAll()
			case <-c.stopCh:
				return
			}
		}
	}()
}

// Stop halts the background checker.
func (c *Checker) Stop() { close(c.stopCh) }

// Overall returns the aggregated health of all targets.
func (c *Checker) Overall() domain.OverallHealth {
	c.mu.RLock()
	defer c.mu.RUnlock()

	overall := domain.StatusHealthy
	targets := make(map[string]domain.TargetHealth, len(c.state))
	for name, th := range c.state {
		targets[name] = *th
		if th.Status != domain.StatusHealthy {
			overall = domain.StatusUnhealthy
		}
	}
	if len(c.state) == 0 {
		overall = domain.StatusHealthy // no targets = trivially healthy
	}
	return domain.OverallHealth{
		Status:    overall,
		Targets:   targets,
		CheckedAt: time.Now().UTC(),
	}
}

// Get returns the health of a single target by name.
func (c *Checker) Get(name string) (domain.TargetHealth, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	th, ok := c.state[name]
	if !ok {
		return domain.TargetHealth{}, false
	}
	return *th, true
}

// Probe performs a single health check against the given URL and returns the result.
// This is exported so tests can call it directly without needing live targets.
func (c *Checker) Probe(ctx context.Context, name, url string) domain.TargetHealth {
	start := time.Now()
	th := domain.TargetHealth{
		Name:        name,
		URL:         url,
		LastChecked: start.UTC(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		th.Status = domain.StatusUnhealthy
		th.Message = fmt.Sprintf("request build error: %v", err)
		return th
	}

	resp, err := c.client.Do(req)
	th.Latency = time.Since(start)

	if err != nil {
		th.Status = domain.StatusUnhealthy
		th.Message = err.Error()
		return th
	}
	defer resp.Body.Close()

	th.StatusCode = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		th.Status = domain.StatusHealthy
	} else {
		th.Status = domain.StatusUnhealthy
		th.Message = fmt.Sprintf("unexpected status %d", resp.StatusCode)
	}
	return th
}

func (c *Checker) runAll() {
	var wg sync.WaitGroup
	for _, t := range c.cfg.Targets {
		wg.Add(1)
		go func(t config.Target) {
			defer wg.Done()
			timeout := t.Timeout
			if timeout == 0 {
				timeout = c.cfg.CheckTimeout
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			result := c.Probe(ctx, t.Name, t.URL)

			c.mu.Lock()
			prev := c.state[t.Name]
			if result.Status == domain.StatusUnhealthy {
				result.Failures = prev.Failures + 1
				if result.Failures < c.cfg.UnhealthyThresh {
					result.Status = domain.StatusHealthy // grace period
				}
			} else {
				result.Failures = 0
			}
			c.state[t.Name] = &result
			c.mu.Unlock()

			c.log.Debug("health probe",
				zap.String("target", t.Name),
				zap.String("status", string(result.Status)),
				zap.Duration("latency", result.Latency),
			)
		}(t)
	}
	wg.Wait()
}
