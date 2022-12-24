package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/stevenferrer/otel-example/telemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	otelCleanup, err := telemetry.Init(ctx, "otel-example-client")
	if err != nil {
		log.Fatalf("init telemetry: %v", err)
	}
	defer otelCleanup()

	otelExampleServerAddr, ok := os.LookupEnv("SERVER_ADDR")
	if !ok {
		log.Fatal("server address not found")
	}

	client := otelExampleClient{
		serverAddr: otelExampleServerAddr,
		httpClient: http.DefaultClient,
	}

	ticker := time.NewTicker(time.Second)
	quit := make(chan bool)
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	go func() {
		for {
			select {
			case <-quit:
				return
			case _ = <-ticker.C:
				client.callEndpoint("/")
			}
		}
	}()

	<-sigChan

	quit <- true
	log.Println("done.")
}

type otelExampleClient struct {
	serverAddr string
	httpClient *http.Client
}

func (c otelExampleClient) callEndpoint(path string) error {
	if path == "" {
		path = "/"
	}

	_, err := c.httpClient.Get(c.serverAddr + path)
	if err != nil {
		return fmt.Errorf("get %s: %w", path, err)
	}

	return nil
}
