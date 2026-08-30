package retry

import "net/http"

type Transport struct {
	Transport http.RoundTripper
}

func (t Transport) Do(req *http.Request) (*http.Response, error) {
	return t.Transport.RoundTrip(req)
}
