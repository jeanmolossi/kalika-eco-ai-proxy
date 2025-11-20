package httpx

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type HTTPMetrics struct {
	requests  metric.Int64Counter
	latencyMs metric.Float64Histogram
	inFlight  metric.Int64UpDownCounter
}

// NewHTTPMetrics cria instruments de métricas HTTP.
// Se meter for nil, retorna nil e não quebra.
func NewHTTPMetrics(meterProvider metric.MeterProvider) (*HTTPMetrics, error) {
	if meterProvider == nil {
		return nil, nil
	}

	meter := meterProvider.Meter("kalika-eco-ai-proxy/http")

	reqCounter, err := meter.Int64Counter(
		"http_server_requests_total",
		metric.WithDescription("Total HTTP requests processed."),
	)
	if err != nil {
		return nil, err
	}

	latencyHist, err := meter.Float64Histogram(
		"http_server_request_duration_ms",
		metric.WithDescription("HTTP request duration in milliseconds."),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return nil, err
	}

	inFlight, err := meter.Int64UpDownCounter(
		"http_server_inflight_requests",
		metric.WithDescription("In-flight HTTP requests."),
	)
	if err != nil {
		return nil, err
	}

	return &HTTPMetrics{
		requests:  reqCounter,
		latencyMs: latencyHist,
		inFlight:  inFlight,
	}, nil
}

// MetricsMiddleware registra métricas HTTP via OpenTelemetry.
// Se m == nil, vira no-op.
func MetricsMiddleware(metrics *HTTPMetrics) echo.MiddlewareFunc {
	if metrics == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			ctx := req.Context()

			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}

			attrs := []attribute.KeyValue{
				attribute.String("http.method", req.Method),
				attribute.String("http.route", route),
				attribute.String("http.target", req.URL.RequestURI()),
				attribute.String("http.host", req.Host),
			}

			start := time.Now()

			metrics.inFlight.Add(ctx, 1, metric.WithAttributes(attrs...))

			defer func() {
				latency := float64(time.Since(start).Milliseconds())

				status := res.Status
				statusClass := status / 100

				allAttrs := append(attrs,
					attribute.Int("http.status_code", status),
					attribute.Int("http.status_class", statusClass),
				)

				metrics.inFlight.Add(ctx, -1, metric.WithAttributes(allAttrs...))
				metrics.requests.Add(ctx, 1, metric.WithAttributes(allAttrs...))
				metrics.latencyMs.Record(ctx, latency, metric.WithAttributes(allAttrs...))
			}()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			return err
		}
	}
}
