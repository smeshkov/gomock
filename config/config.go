package config

import (
	"io/ioutil"
	"time"

	"bitbucket.org/atlassian/sink/logger"
	"gopkg.in/yaml.v2"
)

// Logger is an alias for the external logger type.
type Logger = logger.Logger

var (
	// Log is the global logger.
	Log = logger.NewLogger()
)

// Config represents configuration of application.
type Config struct {
	Server struct {
		Addr         string
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}
	Logger struct {
		Level string
	}
}

// NewConfig loads configuration from file.
func NewConfig(file string) (cfg Config, err error) {

	// Server ...
	cfg.Server.Addr = ":8080"
	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 5 * time.Second
	cfg.Server.IdleTimeout = 5 * time.Second

	// Logger ...
	cfg.Logger.Level = "info"

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	if err = yaml.Unmarshal(data, &cfg); err != nil {
		return
	}

	return
}
