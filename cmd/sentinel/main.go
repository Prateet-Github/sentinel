package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/dataplane"
	"github.com/Prateet-Github/sentinel/internal/proxy"
	"github.com/Prateet-Github/sentinel/internal/router"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewTrieRouter(cfg)

	// p, err := proxy.New(cfg.Backends[0].URL)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	registry, err := proxy.NewRegistry(cfg)
	if err != nil {
		log.Fatal(err)
	}

	dp := dataplane.New(
		r,
		registry,
		cfg,
	)

	log.Printf("Sentinel listening on :%d", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(":8080", dp))
}
