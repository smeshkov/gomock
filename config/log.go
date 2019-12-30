package config

import (
	"fmt"

	"go.uber.org/zap"
)

// SetupLog configures logger.
func SetupLog(level string) {
	l, err := newLog(level)
	if err != nil {
		fmt.Printf("failed to setup logger: %v\n", err)
		return
	}
	zap.ReplaceGlobals(l)
	zap.L().Info("logger is ready")
}

// SyncLog flushes any buffered log entries.
func SyncLog() {
	_ = zap.L().Sync()
}

func newLog(level string) (*zap.Logger, error) {
	if level != "info" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}
