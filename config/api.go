package config

import (
	"encoding/json"
	"io/ioutil"
)

// API represents configuration of API.
type API struct {
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

// NewAPI loads API configuration from file.
func NewAPI(file string) (api API, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &api)
	if err != nil {
		return
	}

	Log.Debug("API configuration:\n%#v", &api)

	return
}
