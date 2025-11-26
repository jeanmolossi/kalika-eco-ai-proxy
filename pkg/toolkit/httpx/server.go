package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/logger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/otel"
)

type Config struct {
	Host                string
	Port                int
	BasePath            string
	ReadTimeout         time.Duration
	ReadHeaderTimeout   time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	MaxRequestBodyBytes int64
	EnablePprof         bool
	EnableTLS           bool
	TLSCertFile         string
	TLSKeyFile          string

	ClientCAsPEM string

	AllowedOrigins []string
}

func Start(cfg Config) func(ctx context.Context, e *echo.Echo) func(context.Context) error {
	if cfg.AllowedOrigins == nil {
		cfg.AllowedOrigins = []string{}
	}

	cfg.BasePath = config.NormalizeBasePath(cfg.BasePath)

	allowedMethods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPatch,
		http.MethodPut,
		http.MethodDelete,
		http.MethodOptions,
	}

	if cfg.MaxRequestBodyBytes <= 0 {
		cfg.MaxRequestBodyBytes = 20 << 20 // 20MiB
	}

	return func(ctx context.Context, e *echo.Echo) func(context.Context) error {
		log := logger.New()

		e.HideBanner = true
		e.HidePort = true

		// Base middlewares
		e.Pre(
			middleware.RemoveTrailingSlash(),
		)

		var middlewares []echo.MiddlewareFunc

		middlewares = append(middlewares,
			RequestID(),
			middleware.Recover(),
			middleware.Secure(),
			middleware.Gzip(),
			middleware.BodyLimit(fmt.Sprintf("%d", cfg.MaxRequestBodyBytes)),
			middleware.CORSWithConfig(middleware.CORSConfig{
				AllowOrigins: cfg.AllowedOrigins,
				AllowMethods: allowedMethods,
			}),
			middleware.TimeoutWithConfig(middleware.TimeoutConfig{
				Timeout: 30 * time.Second,
			}),
		)

		if true {
			tracer := otel.Tracer("kalika-echo-ai-proxy/http")

			httpMetrics, err := NewHTTPMetrics(otel.GetMeterProvider())
			if err != nil {
				log.ErrorContext(ctx, "failed to init http metrics", "err", err)
			}

			middlewares = append(middlewares,
				TraceMiddleware(tracer),
				MetricsMiddleware(httpMetrics),
			)
		}

		middlewares = append(middlewares, LoggerMiddleware(log))

		e.Use(middlewares...)

		baseGroup := e.Group(cfg.BasePath)
		registerInfraRoutes(baseGroup, cfg)

		// BasePath (versioning or gateway-fiendly)
		root := http.NewServeMux()
		root.Handle(cfg.BasePath+"/", e)
		handler := http.TimeoutHandler(root, cfg.WriteTimeout+cfg.ReadTimeout, "server timeout\n")

		srv := &http.Server{
			Addr:              net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
			Handler:           handler,
			ReadTimeout:       cfg.ReadTimeout,
			ReadHeaderTimeout: cfg.ReadHeaderTimeout,
			WriteTimeout:      cfg.WriteTimeout,
			IdleTimeout:       cfg.IdleTimeout,
		}

		// optional TLS (auto HTTP/2 in Go when TLS on)
		if cfg.EnableTLS {
			tlsCfg := &tls.Config{
				MinVersion: tls.VersionTLS12,
				// PreferServerCipherSuites deprecated, not set.
				// ClientAuth is need mTLS
			}

			// if cfg.ClientCAsPEM != "" {
			// 	 load client CAs if need mTLS
			// }

			srv.TLSConfig = tlsCfg
		}

		errCh := make(chan error, 1)

		go func() {
			if cfg.EnableTLS {
				errCh <- srv.ListenAndServeTLS(cfg.TLSCertFile, cfg.TLSKeyFile)
				return
			}

			errCh <- srv.ListenAndServe()
		}()

		startedAt := time.Now()

		log.InfoContext(ctx, "HTTP server listening on "+srv.Addr,
			slog.String("base_path", cfg.BasePath),
			slog.Bool("tls", cfg.EnableTLS),
		)

		return func(ctx context.Context) error {
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

			defer func() {
				signal.Stop(sigCh)
				close(sigCh)
			}()

			select {
			case sig := <-sigCh:
				e.Logger.Infof("shutdown signal received: %v", sig)
			case err := <-errCh:
				if !errors.Is(err, http.ErrServerClosed) {
					return fmt.Errorf("server error: %w", err)
				}
			}

			// timeout context to graceful shutdown
			ctx, cancel := context.WithTimeout(ctx, cfg.ShutdownTimeout)
			defer cancel()

			// try stop accepting new connections, draining keep-alives
			if err := srv.Shutdown(ctx); err != nil {
				e.Logger.Errorf("graceful shutdown failure: %v", err)
				// if stuck, force close
				if err := srv.Close(); err != nil {
					e.Logger.Errorf("forced close failure: %v", err)
				}
			}

			e.Logger.Infof("Server stopped (uptime=%s)", time.Since(startedAt).Truncate(time.Millisecond))

			return nil
		}
	}
}

func registerInfraRoutes(g *echo.Group, cfg Config) {
	g.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	if cfg.EnablePprof {
		g.Any("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))
	}
}
