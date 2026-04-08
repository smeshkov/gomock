// Package main is the entrypoint for the gomock server.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/fsnotify.v1"

	"github.com/smeshkov/gomock/app"
	"github.com/smeshkov/gomock/config"
)

const shutdownTimeout = 5 * time.Second

var version = "untagged"

func main() {
	mockFile := flag.String("mock", "mock.json", "Mock configuration file")
	verbose := flag.Bool("verbose", false, "Verbose")
	ver := flag.Bool("version", false, "prints version of gomock")
	watch := flag.Bool("watch", false, "Watch config file changes and reload automatically")
	flagPort := flag.Int("port", 0, "Server port (overrides mock config)")
	flagAddr := flag.String("addr", "", "Server address e.g. :3000 (overrides port and mock config)")
	flagLogLevel := flag.String("log-level", "", "Log level: info or debug (overrides mock config)")
	flagReadTimeout := flag.String("read-timeout", "", "Read timeout as Go duration e.g. 10s (overrides mock config)")
	flagWriteTimeout := flag.String("write-timeout", "", "Write timeout as Go duration e.g. 10s (overrides mock config)")
	flagIdleTimeout := flag.String("idle-timeout", "", "Idle timeout as Go duration e.g. 60s (overrides mock config)")

	flag.Parse()

	if *ver {
		_, _ = fmt.Fprintln(os.Stdout, version)

		return
	}

	config.SetupLog("info")

	overrides := config.CLIOverrides{
		Port:         *flagPort,
		Addr:         *flagAddr,
		LogLevel:     *flagLogLevel,
		Verbose:      *verbose,
		ReadTimeout:  *flagReadTimeout,
		WriteTimeout: *flagWriteTimeout,
		IdleTimeout:  *flagIdleTimeout,
	}

	serverLoop(*mockFile, *watch, overrides)
}

func serverLoop(mockFile string, watch bool, overrides config.CLIOverrides) {
	// Channel to handle termination signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for {
		mck, mockPath, err := config.NewMock(mockFile)
		if err != nil {
			slog.Warn(fmt.Sprintf("failed to load mock configuration %s: %v", mockFile, err))
		}

		cfg := mck.ToConfig()
		cfg.ApplyOverrides(overrides)
		config.SetupLog(cfg.Logger.Level)

		// Start and monitor server.
		serverCtx, cancelServer := context.WithCancel(context.Background())

		var waitGroup sync.WaitGroup

		waitGroup.Go(func() {
			runServer(serverCtx, &cfg, &mck, version, mockPath)
		})

		if watch {
			go watchConfigFiles(mockFile, cancelServer)
		}

		select {
		case <-sigChan:
			slog.Info("received termination signal, shutting down...")
			cancelServer()
			waitGroup.Wait()

			return
		case <-serverCtx.Done():
			slog.Info("restarting server due to configuration change...")
			waitGroup.Wait()
		}
	}
}

// runServer starts the HTTP server and blocks until ctx is cancelled.
func runServer(ctx context.Context, cfg *config.Config, mck *config.Mock, ver, mockPath string) {
	srv := &http.Server{
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Addr:              cfg.Server.Addr,
		Handler:           app.RegisterHandlers(ver, mockPath, cfg, mck),
	}

	go func() {
		slog.Info(fmt.Sprintf("starting Gomock server at %s (read timeout %s, write timeout %s)",
			cfg.Server.Addr, cfg.Server.ReadTimeout.String(), cfg.Server.WriteTimeout.String()))

		err := srv.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error(fmt.Sprintf("failed to start server: %v", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)

	err := srv.Shutdown(shutdownCtx)

	cancel()

	if err != nil {
		slog.Error(fmt.Sprintf("server shutdown failed: %v", err))
		os.Exit(1)
	}
}

// watchConfigFiles monitors configuration file changes and restarts server.
func watchConfigFiles(mockPath string, cancelServer context.CancelFunc) {
	err := watchLoop(mockPath, cancelServer)
	if err != nil {
		slog.Error(fmt.Sprintf("watcher failed: %v", err))
		os.Exit(1)
	}
}

func watchLoop(mockPath string, cancelServer context.CancelFunc) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}

	defer func() { _ = watcher.Close() }()

	err = watcher.Add(mockPath)
	if err != nil {
		return fmt.Errorf("watching file %s: %w", mockPath, err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				slog.Info(fmt.Sprintf("file %s changed, restarting server...", event.Name))
				cancelServer() // Trigger server shutdown

				return nil // Exit monitoring loop, server will restart in main loop
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}

			slog.Error(fmt.Sprintf("watcher error: %v", err))
		}
	}
}
