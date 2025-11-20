package httpx

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func TraceMiddleware(tracer trace.Tracer) echo.MiddlewareFunc {
	if tracer == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	propagator := otel.GetTextMapPropagator()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			route := c.Path()
			if route == "" {
				route = req.URL.Path
			}

			start := time.Now()

			ctx, span := tracer.Start(ctx,
				route,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.route", route),
					attribute.String("http.target", req.URL.RequestURI()),
					attribute.String("http.scheme", req.URL.Scheme),
					attribute.String("http.host", req.Host),
					attribute.String("net.peer.ip", c.RealIP()),
					attribute.String("user_agent", req.UserAgent()),
				),
			)
			defer span.End()

			// inject new ctx in the request and echo.Context
			req = req.WithContext(ctx)
			c.SetRequest(req)

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			latency := time.Since(start)

			status := res.Status
			span.SetAttributes(
				attribute.Int("http.status_code", status),
				attribute.Int64("http.response_content_length", res.Size),
				attribute.Int64("http.server_latency_ms", latency.Milliseconds()),
			)

			if status >= 500 {
				span.SetStatus(codes.Error, "server error")
			}

			if err != nil {
				span.RecordError(err)
			}

			return err
		}
	}
}
