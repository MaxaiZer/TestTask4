package http

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"runtime/debug"
	"strconv"
	"test-task/internal/services"
	"time"
)

type Config struct {
	Port                int           `validate:"required,min=1,max=65535"`
	DialTimeout         time.Duration `validate:"required,gt=0"`
	KeepAlive           time.Duration `validate:"required,gt=0"`
	MaxIdleConns        int           `validate:"required,gt=0"`
	MaxIdleConnsPerHost int           `validate:"required,gt=0"`
	IdleConnTimeout     time.Duration `validate:"required,gt=0"`
}

type Balancer interface {
	NextServer() (*services.ServerInfo, error)
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewServer(config Config, balancer Balancer) (*http.Server, error) {

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	mux := http.NewServeMux()

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   config.DialTimeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:        config.MaxIdleConns,
		MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.IdleConnTimeout,
	}

	mux.Handle("/", recoverMiddleware(proxyHandler(transport, balancer)))

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Port),
		Handler: mux,
	}
	return server, nil
}

func proxyHandler(transport *http.Transport, balancer Balancer) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server, err := balancer.NextServer()
		if err != nil {
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			slog.Error("couldn't get server", "error", err)
			return
		}

		server.IncConnections()
		defer server.DecConnections()

		targetURL, err := url.Parse(server.Address())
		if err != nil {
			http.Error(w, "Bad upstream", http.StatusBadGateway)
			slog.Error("bad upstream", "url", targetURL)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Transport = transport

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy error", "server", server.Address(), "error", err)
			server.SetHealthy(false)
			http.Error(w, "Upstream server failure", http.StatusBadGateway)
		}

		start := time.Now()
		responseWriter := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		proxy.ServeHTTP(responseWriter, r)

		slog.Info("HTTP request",
			"address", server.Address(),
			"method", r.Method,
			"url", r.URL.String(),
			"status", responseWriter.statusCode,
			"duration", time.Since(start),
		)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recovered",
					"panic", rec,
					"trace", string(debug.Stack()),
					"path", r.URL.Path,
				)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
