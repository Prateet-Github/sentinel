package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	reverseProxy *httputil.ReverseProxy
}

func New(target string) (*Proxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		reverseProxy: httputil.NewSingleHostReverseProxy(u),
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
