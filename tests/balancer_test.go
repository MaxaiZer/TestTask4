package tests

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"test-task/internal/services"
	"test-task/internal/services/balancers"
	myhttp "test-task/internal/transport/http"
	"testing"
	"time"
)

type mockServer struct {
	server   *httptest.Server
	info     *services.ServerInfo
	requests int
}

func newMockServer(delay time.Duration) *mockServer {
	mock := &mockServer{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		io.WriteString(w, "OK")
		mock.requests++
	}))
	info := services.NewServerInfo(server.URL, "/health")
	info.SetHealthy(true)

	mock.server = server
	mock.info = info
	return mock
}

var defaultConfig = myhttp.Config{
	Port:                8081,
	DialTimeout:         30 * time.Second,
	KeepAlive:           30 * time.Second,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,
}

func BenchmarkRoundRobinBalancer(b *testing.B) {

	server1 := newMockServer(0)
	server2 := newMockServer(0)
	defer server1.server.Close()
	defer server2.server.Close()

	balancer := balancers.NewRoundRobinBalancer([]*services.ServerInfo{server1.info, server2.info})

	srv, err := myhttp.NewServer(defaultConfig, balancer)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	go srv.ListenAndServe()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/", defaultConfig.Port))
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			if resp.StatusCode != 200 {
				b.Fatalf("Request failed: %v", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})

	if math.Abs(float64(server1.requests-server2.requests)) > 1 {
		b.Fatalf("Requests are not balanced properly! server1: %d, server2: %d", server1.requests, server2.requests)
	}
}

func BenchmarkLeastConnectionsBalancer(b *testing.B) {

	server1 := newMockServer(0)
	server2 := newMockServer(0)
	defer server1.server.Close()
	defer server2.server.Close()

	balancer := balancers.NewLeastConnectionsBalancer([]*services.ServerInfo{server1.info, server2.info})

	srv, err := myhttp.NewServer(defaultConfig, balancer)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	go srv.ListenAndServe()
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)

	client := &http.Client{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get(fmt.Sprintf("http://localhost:%d/", defaultConfig.Port))
			if err != nil {
				b.Fatalf("Request failed: %v", err)
			}
			if resp.StatusCode != 200 {
				b.Fatalf("Request failed: %v", resp.StatusCode)
			}
			resp.Body.Close()
		}
	})
}
