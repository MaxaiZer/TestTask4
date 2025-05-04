package services

import (
	"log/slog"
	"sync/atomic"
)

type ServerInfo struct {
	address        string
	healthPath     string
	healthy        atomic.Bool
	activeRequests atomic.Int32
}

func NewServerInfo(address string, healthPath string) *ServerInfo {
	s := &ServerInfo{address: address, healthPath: healthPath}
	s.SetHealthy(true)
	return s
}

func (s *ServerInfo) Address() string {
	return s.address
}

func (s *ServerInfo) HealthCheckAddress() string {
	return s.address + s.healthPath
}

func (s *ServerInfo) IsHealthy() bool {
	return s.healthy.Load()
}

func (s *ServerInfo) SetHealthy(value bool) {
	s.healthy.Store(value)
}

func (s *ServerInfo) IncConnections() {
	value := s.activeRequests.Add(1)
	slog.Debug("incremented active requests", "address", s.address, "count", value)
}

func (s *ServerInfo) DecConnections() {
	value := s.activeRequests.Add(-1)
	slog.Debug("decremented active requests", "address", s.address, "count", value)
}

func (s *ServerInfo) Connections() int32 {
	return s.activeRequests.Load()
}
