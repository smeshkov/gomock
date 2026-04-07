package config

import (
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
