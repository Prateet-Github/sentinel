package dataplane

import "net/http"

func (p *Dataplane) HealthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK\n"))
}

func (p *Dataplane) ReadyHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if p.runtimeConfig.Load() == nil {
		http.Error(
			w,
			"not ready\n",
			http.StatusServiceUnavailable,
		)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK\n"))
}
