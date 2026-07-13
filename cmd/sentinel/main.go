package main

import (
	"fmt"
	"log"

	"github.com/Prateet-Github/sentinel/internal/config"
)

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting Sentinel on :%d\n", cfg.Server.Port)
}
