package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/smeshkov/gomock/app"
	c "github.com/smeshkov/gomock/config"
)

var (
	version = "untagged"
)

func main() {
	config := flag.String("config", "_resources/config.yml", "Configuration file")
	mock := flag.String("mock", "mock.json", "Mock configuration file")

	flag.Parse()

	var cfg c.Config
	var mck c.Mock
	var err error

	cfg, err = c.NewConfig(*config)
	if err != nil {
		c.Log.Warn("failed to load configuration %s: %v", *config, err)
	}

	mck, err = c.NewMock(*mock)
	if err != nil {
		c.Log.Fatal("failed to load API configuration %s: %v", *mock, err)
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
		Handler:           app.RegisterHandlers(version, &cfg, &mck),
	}

	c.Log.Info("starting app on %s (read timeout %v, write timeout %v)",
		cfg.Server.Addr, cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)
	if err = srv.ListenAndServe(); err != nil {
		c.Log.Fatal("failed to start server: %v", err)
	}
}
