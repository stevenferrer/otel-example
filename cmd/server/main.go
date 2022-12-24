package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/stevenferrer/otel-example/telemetry"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = 8080
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	otelCleanup, err := telemetry.Init(ctx, "otel-example-server")
	if err != nil {
		log.Fatalf("init telemetry: %v", err)
	}
	defer otelCleanup()

	addr := fmt.Sprintf("%s:%d", getHost(), getPort())
	srv := &http.Server{
		Addr:    addr,
		Handler: newMux(),
	}

	if err := srv.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}

func getHost() string {
	host, ok := os.LookupEnv("SERVER_HOST")
	if !ok {
		return defaultHost
	}

	return host
}

func getPort() int {
	portStr, ok := os.LookupEnv("SERVER_PORT")
	if !ok {
		return defaultPort
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("parse port %v: %v", portStr, err)
	}

	return port
}

func newMux() http.Handler {
	mux := chi.NewMux()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)

	mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	const (
		minDelay = 1
		maxDelay = 10
	)
	mux.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		delay := rand.Intn(maxDelay-minDelay) + minDelay
		time.Sleep(time.Duration(delay) * time.Second)
		w.Write([]byte("ok"))
	})

	return mux
}
