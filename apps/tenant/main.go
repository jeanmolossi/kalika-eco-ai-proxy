package main

import (
	"context"
	"log/slog"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/apps/tenant/runtime"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/core"
	toolkitconfig "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/httpx"
	toolkitlogger "github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/toolkit/logger"
)

func main() {
	ctx := context.Background()
	cfg := toolkitconfig.Load()
	log := toolkitlogger.New().With("service", "tenant")
	app := core.NewApp(log)

	defer toolkitlogger.Flush()

	registry := runtime.Registry()

	serverCfg := httpx.FromToolkitConfig(cfg)
	app.StartServer = httpx.Start(serverCfg)

	if err := app.Start(ctx, core.StartOptions{Registry: registry, Config: cfg}); err != nil {
		log.ErrorContext(ctx, "failed", slog.Any("error", err))
		return
	}
}
