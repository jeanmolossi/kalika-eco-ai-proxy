package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/apps/observability/runtime"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	toolkitlogger "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/logger"
)

func main() {
	ctx := context.Background()
	cfg := toolkitconfig.Load()
	log := toolkitlogger.New().With("service", "observability")
	app := core.NewApp(log)
	cfg.PgDB = toolkitconfig.ChoosePostgres(cfg.ObserveDB, cfg.PgDB)

	defer toolkitlogger.Flush()

	registry := runtime.Registry()

	if cfg.Services.AuthToken != "" {
		app.E.Use(httpx.ServiceAuthMiddleware(cfg.Services.AuthToken))
	}

	cfg.Server = cfg.Observe

	serverCfg := httpx.FromHTTPServer(cfg.Observe)
	app.StartServer = httpx.Start(serverCfg)

	if err := app.Start(ctx, core.StartOptions{Registry: registry, Config: cfg}); err != nil {
		log.ErrorContext(ctx, "failed", slog.Any("error", err))
		toolkitlogger.Flush()
		os.Exit(1)
	}
}
