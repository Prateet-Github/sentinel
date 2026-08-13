package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/Prateet-Github/sentinel/internal/config"
	"github.com/Prateet-Github/sentinel/internal/dataplane"
	"github.com/Prateet-Github/sentinel/internal/lb"
	"github.com/Prateet-Github/sentinel/internal/router"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRadixRouter(cfg)

	loadBalancer := lb.BuildLoadBalancer(cfg)

	dp := dataplane.New(
		r,
		loadBalancer,
		cfg,
	)

	log.Printf("Sentinel listening on :%d", cfg.Server.Port)

	log.Fatal(http.ListenAndServe(":8080", dp))
}
