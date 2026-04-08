package config

import (
	"strconv"
	"time"
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

// CLIOverrides holds CLI flag values that override config settings.
type CLIOverrides struct {
	Port         int
	Addr         string
	LogLevel     string
	Verbose      bool
	ReadTimeout  string
	WriteTimeout string
	IdleTimeout  string
}

// ApplyOverrides applies CLI flag overrides to the config.
func (c *Config) ApplyOverrides(o CLIOverrides) {
	if o.Port > 0 {
		c.Server.Addr = ":" + strconv.Itoa(o.Port)
	}
	if o.Addr != "" {
		c.Server.Addr = o.Addr
	}
	if d, err := time.ParseDuration(o.ReadTimeout); err == nil {
		c.Server.ReadTimeout = d
	}
	if d, err := time.ParseDuration(o.WriteTimeout); err == nil {
		c.Server.WriteTimeout = d
	}
	if d, err := time.ParseDuration(o.IdleTimeout); err == nil {
		c.Server.IdleTimeout = d
	}
	if o.LogLevel != "" {
		c.Logger.Level = o.LogLevel
	}
	if o.Verbose {
		c.Logger.Level = "debug"
	}
}
