package main

import (
	"log"
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/proxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	p, err := proxy.New(cfg.Backends[0].URL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Sentinel listening on :%d", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(":8080", p))
}
