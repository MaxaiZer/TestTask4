package services

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type HealthChecker struct {
	servers             []*ServerInfo
	healthCheckTimeout  time.Duration
	healthCheckInterval time.Duration
	healthTicker        *time.Ticker
	done                chan struct{}
}

func NewHealthChecker(servers []*ServerInfo, healthCheckInterval time.Duration, healthCheckTimeout time.Duration) *HealthChecker {
	return &HealthChecker{
		servers:             servers,
		healthCheckInterval: healthCheckInterval,
		healthCheckTimeout:  healthCheckTimeout,
		done:                make(chan struct{}, 1),
	}
}

func (h *HealthChecker) Run(ctx context.Context) {

	h.healthTicker = time.NewTicker(h.healthCheckInterval)
	client := http.Client{Timeout: h.healthCheckTimeout}

	for {
		select {
		case <-ctx.Done():
			slog.Info("health check stopped")
			h.done <- struct{}{}
			return

		case <-h.healthTicker.C:
			var failed atomic.Int32
			var wg sync.WaitGroup

			slog.Info("health check started")
			for i := range h.servers {
				wg.Add(1)
				go func(s *ServerInfo) {
					defer wg.Done()
					healthy := checkHealth(client, s)
					s.SetHealthy(healthy)
					if !healthy {
						failed.Add(1)
					}
				}(h.servers[i])
			}

			wg.Wait()
			slog.Info("health check finished",
				"healthy", len(h.servers)-int(failed.Load()),
				"unhealthy", failed.Load())
		}
	}
}

func (h *HealthChecker) WaitForStop() {
	<-h.done
}

func checkHealth(client http.Client, server *ServerInfo) bool {
	resp, err := client.Get(server.HealthCheckAddress())
	if err != nil {
		slog.Info("health check failed", "path", server.HealthCheckAddress(), "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Info("health check failed", "path", server.HealthCheckAddress(), "status", resp.Status)
		return false
	}

	return true
}
