package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/httpx"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	log := logger.New()
	app := core.NewApp(log)

	defer logger.Flush()

	registry := core.NewRegistry()

	app.StartServer = httpx.Start(httpx.Config{
		Host:                cfg.Server.Host,
		Port:                cfg.Server.Port,
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
