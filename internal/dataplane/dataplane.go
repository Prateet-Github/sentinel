package dataplane

import (
	"net/http"

	"github.com/Prateet-Github/sentinel/internal/core"
	"github.com/Prateet-Github/sentinel/internal/router"
)

type Dataplane struct {
	router router.Router
	config *core.Config
	proxy  http.Handler
}

func New(
	router router.Router,
	proxy http.Handler,
	config *core.Config,
) *Dataplane {
	return &Dataplane{
		router: router,
		proxy:  proxy,
		config: config,
	}
}

func (p *Dataplane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route, ok := p.router.Match(r.Method, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	_ = route // TODO: will use it later for logging and metrics

	p.proxy.ServeHTTP(w, r)
}
