package config

import (
	"log/slog"
	"os"
	"sync"
)

var (
	levelVar   slog.LevelVar
	loggerOnce sync.Once
)

// SetupLog configures the global slog logger with the given level.
func SetupLog(level string) {
	lvl := parseLevel(level)
	levelVar.Set(lvl)

	loggerOnce.Do(func() {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: &levelVar,
		})

		slog.SetDefault(slog.New(handler))
		slog.Info("logger initialized", "level", level)
	})
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
