package balancers_test

import (
	"github.com/stretchr/testify/assert"
	"test-task/internal/services"
	"test-task/internal/services/balancers"
	"testing"
)

func TestRoundRobinNextServer(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	b := balancers.NewRoundRobinBalancer(servers)

	res, err := b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[0])

	res, err = b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[1])

	res, err = b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[0])
}

func TestRoundRobinNextServer_WithUnhealthy(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	servers[1].SetHealthy(false)

	b := balancers.NewRoundRobinBalancer(servers)

	res, err := b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[0])

	res, err = b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[2])

	res, err = b.NextServer()
	assert.NoError(t, err)
	assert.True(t, res == servers[0])
}

func TestRoundRobinNextServer_NoHealthyServers(t *testing.T) {
	servers := []*services.ServerInfo{
		services.NewServerInfo("addr1", "/health"),
		services.NewServerInfo("addr2", "/health"),
	}

	servers[0].SetHealthy(false)
	servers[1].SetHealthy(false)

	b := balancers.NewRoundRobinBalancer(servers)

	_, err := b.NextServer()
	assert.Error(t, err)
}

func TestRoundRobinNextServer_NoServers(t *testing.T) {
	b := balancers.NewRoundRobinBalancer([]*services.ServerInfo{})
	_, err := b.NextServer()
	assert.Error(t, err)
}
