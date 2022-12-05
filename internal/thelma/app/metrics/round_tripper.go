package metrics

import "net/http"

// adds a bearer token to every request, so that we can authenticate to the IAP-protected
// prometheus gateway
type bearerRoundTripper struct {
	token string
	inner http.RoundTripper
}

func newHttpClientWithBearerToken(token string) *http.Client {
	transport := bearerRoundTripper{
		token: token,
		inner: http.DefaultTransport,
	}
	return &http.Client{
		Transport: transport,
	}
}

func (rt bearerRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Add("Authorization", "Bearer "+rt.token)
	return rt.inner.RoundTrip(r)
}
