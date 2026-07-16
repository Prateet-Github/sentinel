package dataplane

import (
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/proxy"
	"github.com/Prateet-Github/sentinel/internal/router"
)

type Dataplane struct {
	router   router.Router
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
	}
}

func (p *Dataplane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, ok := p.matchRoute(w, r)
	if !ok {
		return
	}

	proxy, ok := p.resolveBackend(w, route)
	if !ok {
		return
	}

	p.forward(proxy, w, r)
}

func (p *Dataplane) matchRoute(
	w http.ResponseWriter,
	r *http.Request,
) (*core.Route, bool) {

	route, ok := p.router.Match(r.Method, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return nil, false
	}

	return route, true
}

func (p *Dataplane) resolveBackend(
	w http.ResponseWriter,
	route *core.Route,
) (http.Handler, bool) {

	proxy, ok := p.registry.Get(route.Backend)
	if !ok {
		http.Error(w, "backend not found", http.StatusBadGateway)
		return nil, false
	}

	return proxy, true
}

func (p *Dataplane) forward(
	proxy http.Handler,
	w http.ResponseWriter,
	r *http.Request,
) {
	proxy.ServeHTTP(w, r)
}
