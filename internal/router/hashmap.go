package router

import (
	"github.com/Prateet-Github/sentinel/internal/core"
)

type HashMapRouter struct {
	routes map[string]core.Route
}

func NewHashMapRouter(cfg *core.Config) *HashMapRouter {
	r := &HashMapRouter{
		routes: make(map[string]core.Route),
	}

	for _, route := range cfg.Routes {
		key := route.Method + ":" + route.Path
		r.routes[key] = route
	}

	return r
}

func (r *HashMapRouter) Match(method, path string) (*core.Route, bool) {
	key := method + ":" + path

	route, ok := r.routes[key]
	if !ok {
		return nil, false
	}

	return &route, true
}
