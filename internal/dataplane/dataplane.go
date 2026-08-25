package dataplane

import (
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/lb"
	"github.com/Prateet-Github/sentinel/internal/proxy"
	"github.com/Prateet-Github/sentinel/internal/router"
)

type Dataplane struct {
	router router.Router
	lb     *lb.LoadBalancer
}

func New(
	router router.Router,
	lb *lb.LoadBalancer,
	config *core.Config,
) *Dataplane {
	return &Dataplane{
		router: router,
		lb:     lb,
	}
}

func (p *Dataplane) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	var params router.Params

	route, ok := p.matchRoute(w, r, &params)
	if !ok {
		return
	}

	r = r.WithContext(router.WithParams(r.Context(), params)) // after successful route match, add params to request context

	selection, ok := p.resolveBackend(w, route)
	if !ok {
		return
	}

	p.forward(selection, w, r)
}

func (p *Dataplane) matchRoute(
	w http.ResponseWriter,
	r *http.Request,
	params *router.Params,
) (*core.Route, bool) {

	route, ok := p.router.Match(r.Method, r.URL.Path, params)
	if !ok {
		http.NotFound(w, r)
		return nil, false
	}

	return route, true
}

func (p *Dataplane) resolveBackend(
	w http.ResponseWriter,
	route *core.Route,
) (*lb.BackendSelection, bool) {

	pool, ok := p.lb.Get(route.Backend)
	if !ok {
		http.Error(
			w,
			"backend pool not found",
			http.StatusBadGateway,
		)
		return nil, false
	}

	selection := pool.Next()
	if selection == nil {
		http.Error(
			w,
			"no backends available",
			http.StatusBadGateway,
		)
		return nil, false
	}

	return selection, true
}

func (p *Dataplane) forward(
	selection *lb.BackendSelection,
	w http.ResponseWriter,
	r *http.Request,
) {
	handler, err := proxy.New(selection.Backend.URL, nil)
	if err != nil {
		http.Error(
			w,
			"failed to create proxy",
			http.StatusBadGateway,
		)
		return
	}

	handler.ServeHTTP(w, r)
}
