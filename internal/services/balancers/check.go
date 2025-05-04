package balancers

import (
	"log/slog"
	"test-task/internal/services"
)

func checkServer(server *services.ServerInfo) bool {
	if server == nil {
		slog.Error("server is nil")
		return false
	}
	return server.IsHealthy()
}
