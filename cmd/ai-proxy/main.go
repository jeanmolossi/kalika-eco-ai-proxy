package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/modules/aiproxy"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/database"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	log := logger.New()
	app := core.NewApp(log)

	defer logger.Flush()

	registry := core.NewRegistry(
		database.NewModule(),
		platform.NewModule(),
		aiproxy.NewModule(),
	)

	app.StartServer = httpx.Start(httpx.Config{
		Host:                cfg.Server.Host,
		Port:                cfg.Server.Port,
		EnableTLS:           cfg.Server.EnableTLS,
		TLSCertFile:         cfg.Server.TLSCertFile,
		TLSKeyFile:          cfg.Server.TLSKeyFile,
		BasePath:            cfg.Server.BasePath,
		ReadTimeout:         cfg.Server.ReadTimeout,
		ReadHeaderTimeout:   2 * time.Second,
		WriteTimeout:        10 * time.Second,
		IdleTimeout:         30 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		MaxRequestBodyBytes: 1 << 20, // 1MiB
	})

	err := app.Start(ctx, core.StartOptions{
		Registry: registry,
		Config:   cfg,
	})
	if err != nil {
		log.ErrorContext(ctx, "failed", slog.Any("error", err))
		return
	}
}
