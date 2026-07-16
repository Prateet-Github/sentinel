package dataplane

import (
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/proxy"
	"github.com/Prateet-Github/sentinel/internal/router"
)

type Dataplane struct {
	router   router.Router
	config   *core.Config
	registry *proxy.Registry
}

func New(
	router router.Router,
	registry *proxy.Registry,
	config *core.Config,
) *Dataplane {
	return &Dataplane{
		router:   router,
		registry: registry,
		config:   config,
	}
}

func (p *Dataplane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, ok := p.router.Match(r.Method, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	backendProxy, ok := p.registry.Get(route.Backend)
	if !ok {
		http.Error(w, "backend not found", http.StatusBadGateway)
		return
	}

	backendProxy.ServeHTTP(w, r)
}
