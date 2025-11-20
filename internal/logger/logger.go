package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jeanmolossi/kalika-eco-ai-proxy/internal/config"
	"github.com/jeanmolossi/kalika-eco-ai-proxy/pkg/bufferwriter"
)

var (
	New = sync.OnceValue(newLogger)
	out io.WriteCloser
)

func newLogger() *slog.Logger {
	cfg := config.Load()

	out = os.Stdout
	if cfg.Environment.Production() {
		out = bufferwriter.NewBufferedWriter(os.Stdout, bufferwriter.BufferedWriterOptions{
			ChannelSize:   2048,
			BufferSize:    128 << 10,
			FlushInterval: 500 * time.Millisecond,
		})
	}

	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{
		Level:     mapLvlStr(cfg.Log.Level),
		AddSource: true,
	}))

	return logger
}

func Flush() error {
	switch w := out.(type) {
	case io.WriteCloser:
		err := w.Close()
		out = nil
		return err
	default:
		out = nil
		return nil
	}
}

func mapLvlStr(s string) slog.Leveler {
	switch {
	case strings.EqualFold(s, "DEBUG"):
		return slog.LevelDebug
	case strings.EqualFold(s, "WARN"):
		return slog.LevelWarn
	case strings.EqualFold(s, "ERROR"):
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
