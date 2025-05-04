package balancers

import (
	"errors"
	"log/slog"
	"sync"
	"test-task/internal/services"
)

type RoundRobinBalancer struct {
	servers            []*services.ServerInfo
	currentServerIndex int32
	mu                 sync.Mutex
}

func NewRoundRobinBalancer(servers []*services.ServerInfo) *RoundRobinBalancer {
	return &RoundRobinBalancer{servers: servers, currentServerIndex: -1}
}

func (r *RoundRobinBalancer) NextServer() (*services.ServerInfo, error) {

	r.mu.Lock()
	defer r.mu.Unlock()

	totalServers := len(r.servers)
	if totalServers == 0 {
		return nil, errors.New("no servers available")
	}

	for i := 0; i < totalServers; i++ {
		r.currentServerIndex = (r.currentServerIndex + 1) % int32(totalServers)
		server := r.servers[r.currentServerIndex]

		if checkServer(server) {
			slog.Debug("next server was chosen", "address", server.Address(), "index", r.currentServerIndex)
			return server, nil
		}
	}

	return nil, errors.New("no healthy servers found")
}
