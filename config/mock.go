package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Mock represents configuration of API.
type Mock struct {
	Port         int         `json:"port,omitempty"`
	Addr         string      `json:"addr,omitempty"`
	ReadTimeout  string      `json:"readTimeout,omitempty"`
	WriteTimeout string      `json:"writeTimeout,omitempty"`
	IdleTimeout  string      `json:"idleTimeout,omitempty"`
	LogLevel     string      `json:"logLevel,omitempty"`
	Endpoints    []*Endpoint `json:"endpoints"`
}

// ToConfig converts Mock server settings into a Config with sensible defaults.
func (m *Mock) ToConfig() Config {
	var cfg Config

	// Defaults
	cfg.Server.Addr = ":8080"
	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 5 * time.Second
	cfg.Server.IdleTimeout = 5 * time.Second
	cfg.Logger.Level = "info"

	// Apply port (addr takes precedence if both set)
	if m.Port > 0 {
		cfg.Server.Addr = ":" + strconv.Itoa(m.Port)
	}
	if m.Addr != "" {
		cfg.Server.Addr = m.Addr
	}

	if d, err := time.ParseDuration(m.ReadTimeout); err == nil {
		cfg.Server.ReadTimeout = d
	}
	if d, err := time.ParseDuration(m.WriteTimeout); err == nil {
		cfg.Server.WriteTimeout = d
	}
	if d, err := time.ParseDuration(m.IdleTimeout); err == nil {
		cfg.Server.IdleTimeout = d
	}

	if m.LogLevel != "" {
		cfg.Logger.Level = m.LogLevel
	}

	return cfg
}

// Endpoint represents API endpoint configuration.
type Endpoint struct {
	Methods   []string    `json:"methods,omitempty"`
	Status    int         `json:"status,omitempty"`
	Path      string      `json:"path"`
	Delay     int         `json:"delay,omitempty"`
	JSONPath  string      `json:"jsonPath,omitempty"` // path to the JSON file with endpoint
	JSON      any `json:"json,omitempty"`
	Proxy     string      `json:"proxy,omitempty"`
	Static    string      `json:"static,omitempty"` // static file server
	Errors    *Errors     `json:"errors,omitempty"`
	AllowCors []string    `json:"allowCors,omitempty"`
	Dynamic   *struct {
		Write *struct {
			JSON *struct {
				Name  string `json:"name"`  // entity name
				Key   string `json:"key"`   // path/to/a/key to store from an incoming JSON
				Value string `json:"value"` // path/to/a/value to store from an incoming JSON
			} `json:"json,omitempty"`
		} `json:"write,omitempty"`
		Read *struct {
			JSON *struct {
				Name     string `json:"name"`               // entity name
				KeyParam string `json:"keyParam,omitempty"` // key parameter name from the "path"
			} `json:"json,omitempty"`
		} `json:"read,omitempty"`
	} `json:"dynamic,omitempty"`
}

// Errors ...
type Errors struct {
	Sample   float32 `json:"sample,omitempty"`
	Statuses []int   `json:"statuses,omitempty"`
}

// NewMock loads API configuration from file.
func NewMock(file string) (mock Mock, path string, err error) {
	path, err = filepath.Abs(file)
	if err != nil {
		return
	}
	path = filepath.Dir(path)
	data, err := os.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &mock)
	if err != nil {
		return
	}

	return
}
