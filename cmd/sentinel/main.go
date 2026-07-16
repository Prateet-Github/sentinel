package main

import (
	"log"
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/dataplane"
	"github.com/Prateet-Github/sentinel/internal/proxy"
	"github.com/Prateet-Github/sentinel/internal/router"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewHashMapRouter(cfg)

	p, err := proxy.New(cfg.Backends[0].URL)
	if err != nil {
		log.Fatal(err)
	}

	dp := dataplane.New(
		r,
		p,
		cfg,
	)

	log.Printf("Sentinel listening on :%d", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(":8080", dp))
}
