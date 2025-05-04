package balancers

import (
	"errors"
	"log/slog"
	"sync"
	"test-task/internal/services"
)

type LeastConnectionsBalancer struct {
	servers []*services.ServerInfo
	mu      sync.Mutex
}

func NewLeastConnectionsBalancer(servers []*services.ServerInfo) *LeastConnectionsBalancer {
	return &LeastConnectionsBalancer{servers: servers}
}

func (r *LeastConnectionsBalancer) NextServer() (*services.ServerInfo, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	var selected *services.ServerInfo
	var minConn int32 = -1

	for _, server := range r.servers {
		if !checkServer(server) {
			continue
		}

		conn := server.Connections()
		if selected == nil || conn < minConn {
			selected = server
			minConn = conn
		}
	}

	if selected == nil {
		return nil, errors.New("no healthy servers found")
	}

	slog.Debug("next server was chosen", "address", selected.Address(), "connections", minConn)
	return selected, nil
}
