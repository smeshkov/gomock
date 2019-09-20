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
	appConfig := flag.String("config", "_resources/config.yml", "Configuration file")
	apiConfig := flag.String("api", "api.json", "API configuration file")

	flag.Parse()

	var cfg c.Config
	var api c.API
	var err error

	cfg, err = c.NewConfig(*appConfig)
	if err != nil {
		c.Log.Warn("failed to load configuration %s: %v", *appConfig, err)
	}

	api, err = c.NewAPI(*apiConfig)
	if err != nil {
		c.Log.Fatal("failed to load API configuration %s: %v", *apiConfig, err)
	}

	if api.Port > 0 {
		cfg.Server.Addr = fmt.Sprintf(":%d", api.Port)
	}

	srv := &http.Server{
		ReadHeaderTimeout: cfg.Server.ReadTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		Addr:              cfg.Server.Addr,
		Handler:           app.RegisterHandlers(version, &cfg, &api),
	}

	c.Log.Info("starting app on %s (read timeout %v, write timeout %v)",
		cfg.Server.Addr, cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)
	if err = srv.ListenAndServe(); err != nil {
		c.Log.Fatal("failed to start server: %v", err)
	}
}
