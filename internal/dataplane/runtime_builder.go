package dataplane

import (
	"fmt"

	"github.com/Prateet-Github/sentinel/internal/core"
	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

func BuildRuntimeState(
	snapshot *controlv1.ConfigSnapshot,
) (*RuntimeState, error) {
	cfg := &core.Config{}

	for _, service := range snapshot.GetServices() {
		for _, backend := range service.GetBackends() {
			cfg.Backends = append(cfg.Backends, core.Backend{
				Name:            backend.GetName(),
				URL:             backend.GetUrl(),
				HealthCheckPath: backend.GetHealthCheckPath(),
			})
		}
	}

	for _, route := range snapshot.GetRoutes() {
		service := findService(
			snapshot.GetServices(),
			route.GetServiceName(),
		)

		if service == nil {
			return nil, fmt.Errorf(
				"route %s %s references unknown service %q",
				route.GetMethod(),
				route.GetPath(),
				route.GetServiceName(),
			)
		}

		for _, backend := range service.GetBackends() {
			cfg.Routes = append(cfg.Routes, core.Route{
				Method:  route.GetMethod(),
				Path:    route.GetPath(),
				Backend: backend.GetName(),
			})
		}
	}

	return &RuntimeState{
		Config: cfg,
	}, nil
}

func findService(
	services []*controlv1.Service,
	name string,
) *controlv1.Service {
	for _, service := range services {
		if service.GetName() == name {
			return service
		}
	}

	return nil
}
