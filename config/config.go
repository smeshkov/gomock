// Package config handles application and mock configuration.
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
func (c *Config) ApplyOverrides(overrides CLIOverrides) {
	if overrides.Port > 0 {
		c.Server.Addr = ":" + strconv.Itoa(overrides.Port)
	}

	if overrides.Addr != "" {
		c.Server.Addr = overrides.Addr
	}

	readTimeout, err := time.ParseDuration(overrides.ReadTimeout)
	if err == nil {
		c.Server.ReadTimeout = readTimeout
	}

	writeTimeout, err := time.ParseDuration(overrides.WriteTimeout)
	if err == nil {
		c.Server.WriteTimeout = writeTimeout
	}

	idleTimeout, err := time.ParseDuration(overrides.IdleTimeout)
	if err == nil {
		c.Server.IdleTimeout = idleTimeout
	}

	if overrides.LogLevel != "" {
		c.Logger.Level = overrides.LogLevel
	}

	if overrides.Verbose {
		c.Logger.Level = "debug"
	}
}
