package config

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

// Mock represents configuration of API.
type Mock struct {
	Port      int         `json:"port,omitempty"`
	Endpoints []*Endpoint `json:"endpoints"`
}

// Endpoint represents API endpoint configuration.
type Endpoint struct {
	Methods   []string    `json:"methods,omitempty"`
	Status    int         `json:"status,omitempty"`
	Path      string      `json:"path"`
	Delay     int         `json:"delay,omitempty"`
	JSONPath  string      `json:"jsonPath,omitempty"`
	JSON      interface{} `json:"json,omitempty"`
	Proxy     string      `json:"proxy,omitempty"`
	Errors    *Errors     `json:"errors,omitempty"`
	AllowCors []string    `json:"allowCors,omitempty"`
	// Mock     json.RawMessage `json:"mock,omitempty"`
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
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &mock)
	if err != nil {
		return
	}

	return
}
