package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/platform/config"
	"github.com/labstack/echo/v4"
)

// App represents the application itself: DI container, HTTP server, logger, etc.
type App struct {
	C *Container
	E *echo.Echo
	L *slog.Logger

	stopFns []func(context.Context) error

	// StartServer allows customizing how the HTTP server is started.
	// If nil, a default based on echo is used.
	StartServer func(context.Context, *echo.Echo) func(context.Context) error
}

// StartOptions controls how the App is initialized.
type StartOptions struct {
	Registry Registry
	Config   *config.Config

	// Optional hooks to customize the bootstrap.
	BeforeModules func(c *Container) error
	AfterModules  func(c *Container) error
}

// NewApp creates an instance of App with defaults.
func NewApp(logger *slog.Logger) *App {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}

	e := echo.New()

	return &App{
		C: NewContainer(),
		E: e,
		L: logger,
	}
}

// Start performs the complete lifecycle of the app:
// - registers core dependencies in the container
// - executes hooks
// - runs module migrations
// - registers providers in the container
// - registers HTTP routes
// - starts modules (background workers, etc.)
// - starts the HTTP server
// - performs graceful shutdown upon receiving a signal or ctx.Done.
func (a *App) Start(ctx context.Context, opt StartOptions) error {
	if opt.Registry == nil {
		return errors.New("core.App: missing Registry in StartOptions")
	}

	if opt.Config == nil {
		return errors.New("core.App: missing Config in StartOptions")
	}

	// 1) Registers core deps in the container
	a.C.Set(ConfigModule, opt.Config)
	a.C.Set(LoggerModule, a.L)
	a.C.Set(EchoModule, a.E)

	// 2) Optional hook before modules (custom setup)
	if opt.BeforeModules != nil {
		if err := opt.BeforeModules(a.C); err != nil {
			return fmt.Errorf("core.App: BeforeModules error: %w", err)
		}
	}

	// 3) Resolve and sort modules by Weight
	mods := opt.Registry.All()
	sort.SliceStable(mods, func(i, j int) bool {
		return mods[i].Weight() < mods[j].Weight()
	})

	a.L.Info("starting modules", slog.Int("count", len(mods)))

	// 4) Provide (register deps)
	for _, m := range mods {
		a.L.Info("providing deps", slog.String("module", m.Name()))

		if err := m.Provide(ctx, a.C); err != nil {
			return fmt.Errorf("core.App: provide failed for module %s: %w", m.Name(), err)
		}
	}

	// 5) Migrations
	a.L.Info("running migrations")

	conn := a.C.MustGet("database:pgconn").(UnwrapConn)
	if conn == nil {
		return fmt.Errorf("core.App: unable to retrieve db connection <nil>")
	}

	if err := RunAllMigrations(ctx, conn.SQL(), mods); err != nil {
		return fmt.Errorf("core.App: migrations failed: %w", err)
	}

	// 6) Optional hook after modules (already with deps registered)
	if opt.AfterModules != nil {
		if err := opt.AfterModules(a.C); err != nil {
			return fmt.Errorf("core.App: AfterModules error: %w", err)
		}
	}

	// 7) Routes
	basePath := config.NormalizeBasePath(opt.Config.Server.BasePath)
	opt.Config.Server.BasePath = basePath

	baseGroup := a.E.Group(basePath)

	for _, m := range mods {
		a.L.Info("registering routes", slog.String("module", m.Name()))

		if err := m.Routes(baseGroup, a.C); err != nil {
			return fmt.Errorf("core.App: routes failed for module %s: %w", m.Name(), err)
		}
	}

	// 8) Start modules (background workers, consumers, schedulers)
	for _, m := range mods {
		a.L.Info("starting module", slog.String("module", m.Name()))

		stopFn, err := m.Start(ctx, a.C)
		if err != nil {
			return fmt.Errorf("core.App: start failed for module %s: %w", m.Name(), err)
		}

		if stopFn != nil {
			a.stopFns = append(a.stopFns, stopFn)
		}
	}

	// 9) Start HTTP server
	startServer := a.StartServer
	if startServer == nil {
		startServer = defaultStartServer
	}

	httpStopFn := startServer(ctx, a.E)
	if httpStopFn != nil {
		a.stopFns = append(a.stopFns, httpStopFn)
	}

	// 10) Wait for shutdown signal (or ctx.Done)
	// external ctx may come from main with cancel by signal as well.
	sigCtx, stopSignal := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stopSignal()

	a.L.Info("application started, waiting for shutdown signal")

	<-sigCtx.Done()

	a.L.Info("shutdown requested, starting graceful shutdown")

	// 11) Graceful shutdown: executes stopFns in reverse order
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var firstErr error

	for i := len(a.stopFns) - 1; i >= 0; i-- {
		stop := a.stopFns[i]
		if stop == nil {
			continue
		}

		if err := stop(shutdownCtx); err != nil {
			if firstErr == nil {
				firstErr = err
			}

			a.L.Error("error during shutdown", slog.Any("err", err))
		}
	}

	if firstErr != nil {
		a.L.Error("shutdown finished with errors", slog.Any("err", firstErr))
		return firstErr
	}

	a.L.Info("shutdown completed successfully")

	return nil
}

// defaultStartServer starts the echo with a basic listen and returns a shutdown function.
func defaultStartServer(ctx context.Context, e *echo.Echo) func(context.Context) error {
	addr := ":8080"

	// Canal para capturar erro do Start
	errCh := make(chan error, 1)

	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Opcional: você pode ter outra go routine aqui pra logar erro fatal.

	return func(ctx context.Context) error {
		return e.Shutdown(ctx)
	}
}
