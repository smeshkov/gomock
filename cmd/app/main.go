package main

import (
	"flag"
	"fmt"
	"net/http"

	"go.uber.org/zap"

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

	flag.Parse()

	var cfg config.Config
	var err error

	cfg, err = config.NewConfig(*configFile)
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

	if mck.Port > 0 {
		cfg.Server.Addr = fmt.Sprintf(":%d", mck.Port)
	}

	srv := &http.Server{
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Addr:              cfg.Server.Addr,
		Handler:           app.RegisterHandlers(version, mockPath, &cfg, &mck),
	}

	zap.L().Info(fmt.Sprintf("starting proxy app on %s (read timeout %s, write timeout %s)",
		cfg.Server.Addr, cfg.Server.ReadTimeout.String(), cfg.Server.WriteTimeout.String()))
	if err = srv.ListenAndServe(); err != nil {
		zap.L().Fatal(fmt.Sprintf("failed to start server: %v", err))
	}
}
