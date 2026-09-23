package dataplane

import (
	"fmt"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/lb"
	"github.com/Prateet-Github/sentinel/internal/router"
	controlv1 "github.com/Prateet-Github/sentinel/proto"
)

func BuildRuntimeState(
	snapshot *controlv1.ConfigSnapshot,
) (*RuntimeState, error) {
	cfg := &core.Config{}

	// service to backend pool
	for _, service := range snapshot.GetServices() {
		for _, backend := range service.GetBackends() {
			cfg.Backends = append(cfg.Backends, core.Backend{
				Name:            backend.GetName(),
				Service:         service.GetName(),
				URL:             backend.GetUrl(),
				HealthCheckPath: backend.GetHealthCheckPath(),
			})
		}
	}

	// routes to service/backend pool
	for _, route := range snapshot.GetRoutes() {
		if findService(snapshot.GetServices(), route.GetServiceName()) == nil {
			return nil, fmt.Errorf(
				"route %s %s references unknown service %q",
				route.GetMethod(),
				route.GetPath(),
				route.GetServiceName(),
			)
		}

		cfg.Routes = append(cfg.Routes, core.Route{
			Method:  route.GetMethod(),
			Path:    route.GetPath(),
			Backend: route.GetServiceName(),
		})
	}

	runtimeRouter := router.NewRadixRouter(cfg)
	runtimeLoadBalancer := lb.BuildLoadBalancer(cfg)

	return &RuntimeState{
		Config:       cfg,
		Router:       runtimeRouter,
		LoadBalancer: runtimeLoadBalancer,
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
