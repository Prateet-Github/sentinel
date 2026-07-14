package dataplane

import "net/http"

type Dataplane struct {
	proxy http.Handler
}

func New(proxy http.Handler) *Dataplane {
	return &Dataplane{
		proxy: proxy,
	}
}

func (p *Dataplane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
