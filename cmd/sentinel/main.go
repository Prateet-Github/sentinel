package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/dataplane"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	runtimeConfig := dataplane.NewRuntimeConfig()

	dp := dataplane.New(
		runtimeConfig,
		cfg,
	)

	/*
		Control Plane client.
	*/
	controlClient, err := dataplane.NewControlClient(
		ctx,
		"localhost:9090",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer controlClient.Close()

	/*
		Start Control Plane stream.

		The goroutine exits automatically when the root
		context is cancelled.
	*/
	go func() {
		if err := controlClient.StreamConfig(
			ctx,
			"dp-1",
			runtimeConfig,
		); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf(
				"control plane stream ended: %v",
				err,
			)
		}
	}()

	/*
		Do not expose the HTTP dataplane until the first
		valid Control Plane snapshot has been received.
	*/
	log.Println("waiting for initial control plane configuration...")

	if err := runtimeConfig.WaitReady(ctx); err != nil {
		log.Println("shutdown before initial control plane configuration")
		return
	}

	log.Println("initial control plane configuration received")

	/*
		Data Plane HTTP server.
	*/
	mux := http.NewServeMux()

	mux.HandleFunc("/health", dp.HealthHandler)
	mux.HandleFunc("/ready", dp.ReadyHandler)

	mux.Handle("/", dp)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	/*
		pprof server.

		Kept separate from the dataplane server.
	*/
	pprofServer := &http.Server{
		Addr:    "localhost:6060",
		Handler: http.DefaultServeMux,
	}

	/*
		Start pprof.
	*/
	go func() {
		log.Println("pprof listening on localhost:6060")

		if err := pprofServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf(
				"pprof server error: %v",
				err,
			)
		}
	}()

	/*
		Start Data Plane.
	*/
	go func() {
		log.Printf(
			"Sentinel listening on :%d",
			cfg.Server.Port,
		)

		if err := httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			log.Printf(
				"data plane server error: %v",
				err,
			)
		}
	}()

	/*
		Wait for shutdown signal.
	*/
	<-ctx.Done()

	log.Println("shutdown signal received")

	/*
		Give active HTTP requests time to finish.
	*/
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	/*
		Stop accepting new dataplane requests.
	*/
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"data plane shutdown error: %v",
			err,
		)
	}

	/*
		Stop pprof.
	*/
	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		log.Printf(
			"pprof shutdown error: %v",
			err,
		)
	}

	/*
		The root context cancellation has already caused
		the Control Plane stream to stop.

		Close the gRPC connection explicitly.
	*/
	if err := controlClient.Close(); err != nil {
		log.Printf(
			"control client close error: %v",
			err,
		)
	}

	log.Println("Sentinel shutdown complete")
}
