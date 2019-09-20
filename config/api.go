package config

import (
	"encoding/json"
	"io/ioutil"
)

// API represents configuration of API.
type API struct {
	Port      int
	Endpoints []Endpoint
}

// Endpoint represents API endpoint configuration.
type Endpoint struct {
	Method   string
	Status   int
	Path     string
	JSONPath string
	JSON     string
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

	return
}
