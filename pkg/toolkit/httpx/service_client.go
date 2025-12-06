package httpx

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	maigohttpx "github.com/jeanmolossi/maigo/pkg/httpx"
	"github.com/jeanmolossi/maigo/pkg/httpx/circuitbreaker"
	maigocontracts "github.com/jeanmolossi/maigo/pkg/maigo/contracts"
)

// CircuitBreakerConfig defines the thresholds used by the circuit-breaker
// transport for inter-service calls.
type CircuitBreakerConfig struct {
	FailureThreshold int
	RecoveryWindow   time.Duration
	ShouldTrip       func(*http.Response, error) bool

	// Deprecated compatibility fields kept to smooth migration from the previous
	// gobreaker implementation. Prefer using FailureThreshold/RecoveryWindow
	// above instead.
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
	CACertFile string
}

// NewServiceHTTPClient wires retry and circuit-breaker protections for
// inter-service HTTP calls.
func NewServiceHTTPClient(opts ServiceClientOptions) (*http.Client, error) {
	baseTransport := cloneDefaultTransport()

	if opts.CACertFile != "" {
		caPool, err := loadRootCAs(opts.CACertFile)
		if err != nil {
			return nil, fmt.Errorf("load root CAs: %w", err)
		}

		if baseTransport.TLSClientConfig == nil {
			baseTransport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		}

		baseTransport.TLSClientConfig.RootCAs = caPool
	}

	transport := maigohttpx.Compose(baseTransport, circuitbreaker.WithCircuitBreaker(newCircuitBreakerConfig(opts.Breaker)))

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient.Transport = transport
	retryClient.HTTPClient.Timeout = opts.Timeout

	if opts.MaxRetries > 0 {
		retryClient.RetryMax = opts.MaxRetries
	}

	return retryClient.StandardClient(), nil
}

func cloneDefaultTransport() *http.Transport {
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		return base.Clone()
	}

	return &http.Transport{}
}

func loadRootCAs(caFile string) (*x509.CertPool, error) {
	data, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("read ca file %s: %w", caFile, err)
	}

	pool := x509.NewCertPool()

	if ok := pool.AppendCertsFromPEM(data); !ok {
		return nil, fmt.Errorf("no certificates appended from %s", caFile)
	}

	return pool, nil
}

func newCircuitBreakerConfig(cfg CircuitBreakerConfig) circuitbreaker.CircuitBreakerConfig {
	failureThreshold := cfg.FailureThreshold
	if failureThreshold == 0 {
		failureThreshold = int(cfg.Failures)
	}

	if failureThreshold == 0 {
		failureThreshold = 5
	}

	recoveryWindow := cfg.RecoveryWindow
	if recoveryWindow == 0 {
		recoveryWindow = cfg.ResetTimeout
	}

	if recoveryWindow == 0 {
		recoveryWindow = 30 * time.Second
	}

	return circuitbreaker.CircuitBreakerConfig{
		FailureThreshold: failureThreshold,
		RecoveryWindow:   recoveryWindow,
		ShouldTrip:       cfg.ShouldTrip,
	}
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
