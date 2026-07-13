package config

import "github.com/Prateet-Github/sentinel/internal/core"

func Load() (*core.Config, error) {
	// will make a yaml file for this later when the parser is done
	return &core.Config{
		Server: core.Server{
			Port: 8080,
		},
		Backends: []core.Backend{
			{
				Name: "users-service",
				URL:  "http://localhost:9000",
			},
		},
		Routes: []core.Route{
			{
				Method:  "GET",
				Path:    "/users",
				Backend: "users-service",
			},
		},
	}, nil
}
