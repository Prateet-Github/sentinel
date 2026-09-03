package proxy

import (
	"errors"
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
	onResponse   func(*http.Response)
	target       *url.URL
}

func New(
	target string,
	onError func(error),
	onResponse func(*http.Response),
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
		onResponse:   onResponse,
		target:       u,
	}

	rp.ModifyResponse = func(resp *http.Response) error {
		if p.onResponse != nil {
			p.onResponse(resp)
		}

		return nil
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

func (p *Proxy) Attempt(r *http.Request) (*http.Response, error) {
	upstream := r.Clone(r.Context())

	upstream.URL.Scheme = p.target.Scheme
	upstream.URL.Host = p.target.Host

	if r.Body != nil && r.ContentLength != 0 {
		if r.GetBody == nil {
			return nil, errors.New("request body is not replayable")
		}

		body, err := r.GetBody()
		if err != nil {
			return nil, err
		}

		upstream.Body = body
	}

	resp, err := optimizedTransport.RoundTrip(upstream)
	if err != nil {
		if p.onError != nil {
			p.onError(err)
		}
		return nil, err
	}

	if p.onResponse != nil {
		p.onResponse(resp)
	}

	return resp, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.reverseProxy.ServeHTTP(w, r)
}
