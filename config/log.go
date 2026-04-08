package config

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

var (
	logger     *zap.Logger
	loggerLock sync.Mutex
)

// SetupLog - setup function (runs once on initialization, then only changes levels).
func SetupLog(level string) {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if logger == nil {
		// create logger only on first run
		newLogger, err := newLog(level)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to setup logger: %v\n", err)

			return
		}

		logger = newLogger
		zap.ReplaceGlobals(logger) // setting up a global logger
		zap.L().Sugar().Infof("logger initialized with level [%s]", level)
	} else {
		// if you have an existing logger, just change the level
		zap.L().Sugar().Infof("updating logger level to [%s]", level)

		_ = logger.Sync() // flush existing logger buffer
		newLogger, _ := newLog(level)
		logger = newLogger
	}
}

// SyncLog - buffer flush function.
func SyncLog() {
	loggerLock.Lock()
	defer loggerLock.Unlock()

	if logger != nil {
		_ = logger.Sync()
	}
}

// newLog - internal function (create logger).
func newLog(level string) (*zap.Logger, error) {
	if level != "info" {
		zapLogger, err := zap.NewDevelopment()
		if err != nil {
			return nil, fmt.Errorf("creating development logger: %w", err)
		}

		return zapLogger, nil
	}

	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("creating production logger: %w", err)
	}

	return zapLogger, nil
}
