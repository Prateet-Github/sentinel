package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

var optimizedTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).DialContext,

	MaxIdleConns:        10000,
	MaxIdleConnsPerHost: 2000,
	MaxConnsPerHost:     0,

	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	ForceAttemptHTTP2:     false,
}

type Proxy struct {
	reverseProxy *httputil.ReverseProxy
	onError      func(error)
}

func New(
	target string,
	onError func(error),
) (*Proxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Transport = optimizedTransport

	p := &Proxy{
		reverseProxy: rp,
		onError:      onError,
	}

	rp.ErrorHandler = func(
		w http.ResponseWriter,
		r *http.Request,
		err error,
	) {
		if p.onError != nil {
			p.onError(err)
		}

		http.Error(
			w,
			"upstream unavailable",
			http.StatusBadGateway,
		)
	}

	return p, nil
}
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
