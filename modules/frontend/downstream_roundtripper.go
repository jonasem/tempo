package frontend

import (
	"net/http"
	"net/url"
	"path"

	"github.com/opentracing/opentracing-go"
)

// RoundTripper that forwards requests to downstream URL.
type downstreamRoundTripper struct {
	downstreamURL *url.URL
	transport     http.RoundTripper
}

func NewDownstreamRoundTripper(downstreamURL string, transport http.RoundTripper) (http.RoundTripper, error) {
	u, err := url.Parse(downstreamURL)
	if err != nil {
		return nil, err
	}

	return &downstreamRoundTripper{downstreamURL: u, transport: transport}, nil
}

func (d downstreamRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	tracer, span := opentracing.GlobalTracer(), opentracing.SpanFromContext(r.Context())
	if tracer != nil && span != nil {
		carrier := make(opentracing.TextMapCarrier, len(r.Header))
		for k, v := range r.Header {
			carrier.Set(k, v[0])
		}
		err := tracer.Inject(span.Context(), opentracing.TextMap, carrier)
		if err != nil {
			return nil, err
		}
	}

	r.URL.Scheme = d.downstreamURL.Scheme
	r.URL.Host = d.downstreamURL.Host
	r.URL.Path = path.Join(d.downstreamURL.Path, r.URL.Path)
	r.Host = ""
	return d.transport.RoundTrip(r)
}
