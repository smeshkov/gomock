package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/fsnotify.v1"

	"github.com/smeshkov/gomock/app"
	"github.com/smeshkov/gomock/config"
)

var (
	version = "untagged"
)

func main() {
	configFile := flag.String("config", "_resources/config.yml", "Configuration file")
	mockFile := flag.String("mock", "mock.json", "Mock configuration file")
	verbose := flag.Bool("verbose", false, "Verbose")
	ver := flag.Bool("version", false, "prints version of Tagify")
	watch := flag.Bool("watch", false, "Watch config file changes and reload automatically")

	flag.Parse()

	if *ver {
		fmt.Println(version)
		return
	}

	// Initialize logging
	config.SetupLog("info")

	// Channel to handle termination signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Server execution loop
	for {
		// Load initial configuration
		cfg, err := config.NewConfig(*configFile)
		if err != nil {
			zap.L().Warn(fmt.Sprintf("failed to load configuration %s: %v", *configFile, err))
		}

		logLevel := cfg.Logger.Level
		if *verbose {
			logLevel = "debug"
		}

		config.SetupLog(logLevel)

		var mck config.Mock
		var mockPath string
		mck, mockPath, err = config.NewMock(*mockFile)
		if err != nil {
			zap.L().Warn(fmt.Sprintf("failed to load API configuration %s: %v", *mockFile, err))
		}

		// Start and monitor server
		serverCtx, cancelServer := context.WithCancel(context.Background())
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			defer wg.Done()
			runServer(serverCtx, &cfg, &mck, version, mockPath)
		}()

		// Enable configuration file monitoring
		if *watch {
			go watchConfigFiles(*configFile, *mockFile, cancelServer)
		}

		select {
		case <-sigChan:
			// Terminate server upon receiving termination signal
			zap.L().Info("received termination signal, shutting down...")
			cancelServer()
			wg.Wait()
			return
		case <-serverCtx.Done():
			// Restart server upon configuration change
			zap.L().Info("restarting server due to configuration change...")
			wg.Wait()
			// Continue loop to restart server with updated configuration
		}
	}
}

// Server execution function
func runServer(ctx context.Context, cfg *config.Config, mck *config.Mock, version, mockPath string) {
	srv := &http.Server{
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Addr:              cfg.Server.Addr,
		Handler:           app.RegisterHandlers(version, mockPath, cfg, mck),
	}

	go func() {
		zap.L().Info(fmt.Sprintf("starting proxy app on %s (read timeout %s, write timeout %s)",
			cfg.Server.Addr, cfg.Server.ReadTimeout.String(), cfg.Server.WriteTimeout.String()))

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			zap.L().Fatal(fmt.Sprintf("failed to start server: %v", err))
		}
	}()

	<-ctx.Done()
	zap.L().Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		zap.L().Fatal(fmt.Sprintf("server shutdown failed: %v", err))
	}
}

// Monitor configuration file changes and restart server
func watchConfigFiles(configPath, mockPath string, cancelServer context.CancelFunc) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zap.L().Fatal(fmt.Sprintf("failed to create watcher: %v", err))
	}
	defer watcher.Close()

	// Add files to monitor
	files := []string{configPath, mockPath}
	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			zap.L().Fatal(fmt.Sprintf("failed to watch file %s: %v", file, err))
		}
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				zap.L().Info(fmt.Sprintf("file %s changed, restarting server...", event.Name))
				cancelServer() // Trigger server shutdown
				return         // Exit monitoring loop, server will restart in main loop
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			zap.L().Error(fmt.Sprintf("watcher error: %v", err))
		}
	}
}
