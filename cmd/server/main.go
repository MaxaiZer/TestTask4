package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {

	port := "8080"

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(fmt.Sprintf("Hello")))
	})

	http.HandleFunc("/sleep", func(w http.ResponseWriter, r *http.Request) {
		delayStr := r.URL.Query().Get("delay")
		delaySec, err := strconv.Atoi(delayStr)
		if err != nil || delaySec < 0 {
			http.Error(w, "invalid delay", http.StatusBadRequest)
			return
		}

		slog.Info("Sleeping", "seconds", delaySec)
		time.Sleep(time.Duration(delaySec) * time.Second)

		_, _ = w.Write([]byte(fmt.Sprintf("Slept for %d seconds", delaySec)))
	})

	addr := ":" + port
	slog.Info("Server is running")
	if err := http.ListenAndServe(addr, nil); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
