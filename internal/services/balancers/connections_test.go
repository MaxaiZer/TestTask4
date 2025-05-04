package balancers_test

import (
	"github.com/stretchr/testify/assert"
	"test-task/internal/services"
	"test-task/internal/services/balancers"
	"testing"
)

func TestConnectionsNextServer(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	servers[0].IncConnections()
	servers[0].IncConnections()
	servers[1].IncConnections()

	b := balancers.NewLeastConnectionsBalancer(servers)

	res, err := b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[1])
}

func TestConnectionsNextServer_WithUnhealthy(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	servers[0].IncConnections()
	servers[0].IncConnections()
	servers[1].IncConnections()
	servers[1].SetHealthy(false)

	b := balancers.NewLeastConnectionsBalancer(servers)

	res, err := b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[0])
}

func TestConnectionsNextServer_NoHealthyServers(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	servers[0].SetHealthy(false)
	servers[1].SetHealthy(false)

	b := balancers.NewLeastConnectionsBalancer(servers)

	_, err := b.NextServer()
	assert.Error(t, err)
}

func TestConnectionsNextServer_NoServers(t *testing.T) {
	b := balancers.NewLeastConnectionsBalancer([]*services.ServerInfo{})
	_, err := b.NextServer()
	assert.Error(t, err)
}
