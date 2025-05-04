package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	nethttp "net/http"
	"os"
	"test-task/internal/config"
	"test-task/internal/services"
	"test-task/internal/services/balancers"
	"test-task/internal/transport/http"
)

type App struct {
	Config        *config.Config
	server        *nethttp.Server
	checker       *services.HealthChecker
	checkerCancel context.CancelFunc
}

func New() (*App, error) {

	cfg, err := config.Get()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	slog.Info("current environment", "env", cfg.Env)

	initLogger(cfg)

	servers := make([]*services.ServerInfo, len(cfg.Servers))
	for i, s := range cfg.Servers {
		servers[i] = services.NewServerInfo(s.Address, s.HealthPath)
	}

	var balancer http.Balancer
	switch cfg.Algorithm {
	case config.RoundRobin:
		balancer = balancers.NewRoundRobinBalancer(servers)
	case config.LeastConnections:
		balancer = balancers.NewLeastConnectionsBalancer(servers)
	default:
		return nil, errors.New("invalid algorithm")
	}

	checker := services.NewHealthChecker(servers, cfg.HealthCheckInterval, cfg.HealthCheckTimeout)
	server, err := http.NewServer(http.Config{
		Port:                cfg.Port,
		DialTimeout:         cfg.DialTimeout,
		KeepAlive:           cfg.KeepAlive,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout},
		balancer)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	return &App{Config: cfg, server: server, checker: checker}, nil
}

func (a *App) Run() {

	ctx, cancel := context.WithCancel(context.Background())
	a.checkerCancel = cancel
	go a.checker.Run(ctx)

	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

func (a *App) Stop(ctx context.Context) {

	slog.Info("shutting down gracefully...")

	if err := a.server.Shutdown(ctx); err != nil {
		slog.Error("failed to gracefully shutdown server", "error", err)
	} else {
		slog.Info("HTTP server gracefully stopped")
	}

	a.checkerCancel()
	a.checker.WaitForStop()
}

func initLogger(cfg *config.Config) {
	var handler slog.Handler

	if cfg.Env == config.Production {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
