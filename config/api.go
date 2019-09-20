package config

import (
	"encoding/json"
	"io/ioutil"
)

// Mock represents configuration of API.
type Mock struct {
	Port      int         `json:"port,omitempty"`
	Endpoints []*Endpoint `json:"endpoints"`
}

// Endpoint represents API endpoint configuration.
type Endpoint struct {
	Method   string      `json:"method,omitempty"`
	Status   int         `json:"status,omitempty"`
	Path     string      `json:"path"`
	JSONPath string      `json:"jsonPath,omitempty"`
	JSON     interface{} `json:"json,omitempty"`
}

// NewMock loads API configuration from file.
func NewAPI(file string) (mock Mock, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &mock)
	if err != nil {
		return
	}

	Log.Debug("mock configuration: %#v", &mock)

	return
}
