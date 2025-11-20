package httpx

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
)

func LoggerMiddleware(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			ctx := req.Context()

			status := res.Status

			var level slog.Level

			switch {
			case status >= 500 || err != nil:
				level = slog.LevelError
			case status >= 400:
				level = slog.LevelWarn
			default:
				level = slog.LevelInfo
			}

			requestID := res.Header().Get(echo.HeaderXRequestID)
			if requestID == "" {
				requestID = req.Header.Get(echo.HeaderXRequestID)
			}

			attrs := []slog.Attr{
				slog.String("method", req.Method),
				slog.String("route", c.Path()),
				slog.String("uri", req.RequestURI),
				slog.Int("status", status),
				slog.Int64("latency_ms", latency.Milliseconds()),
				slog.Int64("bytes_in", req.ContentLength),
				slog.Int64("bytes_out", res.Size),
				slog.String("remote_ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
				slog.String("host", req.Host),
			}

			if requestID != "" {
				attrs = append(attrs, slog.String("request_id", requestID))
			}

			if err != nil {
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			spanCtx := trace.SpanFromContext(ctx).SpanContext()
			if spanCtx.IsValid() {
				attrs = append(attrs,
					slog.String("trace_id", spanCtx.TraceID().String()),
					slog.String("span_id", spanCtx.SpanID().String()),
				)
			}

			log.LogAttrs(ctx, level, "http_request", attrs...)

			return err
		}
	}
}
