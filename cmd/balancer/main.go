package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"test-task/internal/app"
)

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	myapp, err := app.New()
	if err != nil {
		slog.Error("failed to initialize app", "error", err)
		os.Exit(1)
	}

	go myapp.Run()

	<-stop
	slog.Info("received termination signal")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), myapp.Config.ShutdownTimeout)
	defer cancel()

	myapp.Stop(shutdownCtx)
	slog.Info("application stopped")
}
