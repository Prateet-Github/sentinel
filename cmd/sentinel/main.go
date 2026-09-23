package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/dataplane"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	runtimeConfig := dataplane.NewRuntimeConfig()

	dp := dataplane.New(
		runtimeConfig,
		cfg,
	)

	ctx := context.Background()

	controlClient, err := dataplane.NewControlClient(
		ctx,
		"localhost:9090",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer controlClient.Close()

	go func() {
		if err := controlClient.StreamConfig(
			ctx,
			"dp-1",
			runtimeConfig,
		); err != nil {
			log.Printf("control plane stream ended: %v", err)
		}
	}()

	log.Printf("Sentinel listening on :%d", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(":8080", dp))
}
