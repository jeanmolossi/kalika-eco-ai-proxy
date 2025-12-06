package httpx

import (
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
	"github.com/sony/gobreaker/v2"
)

// CircuitBreakerConfig defines the thresholds used by the circuit-breaker
// transport for inter-service calls.
type CircuitBreakerConfig struct {
	Failures     uint32
	ResetTimeout time.Duration
	Interval     time.Duration
}

// ServiceClientOptions bundles resilience settings for HTTP clients that talk
// to other bounded contexts.
type ServiceClientOptions struct {
	Token      string
	Timeout    time.Duration
	MaxRetries int
	Breaker    CircuitBreakerConfig
}

// NewServiceHTTPClient wires retry and circuit-breaker protections for
// inter-service HTTP calls.
func NewServiceHTTPClient(opts ServiceClientOptions) *http.Client {
	breaker := newCircuitBreaker(opts.Breaker)
	baseTransport := http.DefaultTransport

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = &breakerRoundTripper{next: baseTransport, breaker: breaker}
	retryClient.HTTPClient.Timeout = opts.Timeout

	if opts.MaxRetries > 0 {
		retryClient.RetryMax = opts.MaxRetries
	}

	return retryClient.StandardClient()
}

func newCircuitBreaker(cfg CircuitBreakerConfig) *gobreaker.CircuitBreaker[*http.Response] {
	failures := cfg.Failures
	if failures == 0 {
		failures = 5
	}

	resetTimeout := cfg.ResetTimeout
	if resetTimeout == 0 {
		resetTimeout = 30 * time.Second
	}

	interval := cfg.Interval
	if interval == 0 {
		interval = 2 * time.Minute
	}

	return gobreaker.NewCircuitBreaker[*http.Response](gobreaker.Settings{
		Name:     "service-http",
		Timeout:  resetTimeout,
		Interval: interval,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= failures
		},
	})
}

type breakerRoundTripper struct {
	next    http.RoundTripper
	breaker *gobreaker.CircuitBreaker[*http.Response]
}

func (b *breakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return b.breaker.Execute(func() (*http.Response, error) {
		return b.next.RoundTrip(req)
	})
}

// AsMaigoHTTPClient adapts a standard *http.Client to the maigocontracts.HTTPClientCompat interface.
func AsMaigoHTTPClient(client *http.Client) maigocontracts.HTTPClientCompat {
	return &maigoHTTPClient{client: client}
}

type maigoHTTPClient struct {
	client *http.Client
}

func (m *maigoHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.client.Do(req)
}

func (m *maigoHTTPClient) Transport() http.RoundTripper {
	return m.client.Transport
}

func (m *maigoHTTPClient) SetTransport(rt http.RoundTripper) {
	m.client.Transport = rt
}

func (m *maigoHTTPClient) Timeout() time.Duration {
	return m.client.Timeout
}

func (m *maigoHTTPClient) SetTimeout(d time.Duration) {
	m.client.Timeout = d
}

func (m *maigoHTTPClient) SetFollowRedirects(follow bool) {
	if follow {
		m.client.CheckRedirect = nil
		return
	}

	m.client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
}

func (m *maigoHTTPClient) Unwrap() *http.Client {
	return m.client
}
