package proxy

import (
	"fmt"
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/core"
)

type Registry struct {
	proxies map[string]http.Handler
}

func NewRegistry(cfg *core.Config) (*Registry, error) {
	r := &Registry{
		proxies: make(map[string]http.Handler),
	}

	for _, backend := range cfg.Backends {
		p, err := New(backend.URL, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("proxy %s: %w", backend.Name, err)
		}

		r.proxies[backend.Name] = p
	}

	return r, nil
}

func NewRegistryFromHandlers(
	proxies map[string]http.Handler,
) *Registry {
	return &Registry{
		proxies: proxies,
	}
}

func (r *Registry) Get(name string) (http.Handler, bool) {
	p, ok := r.proxies[name]
	return p, ok
}
