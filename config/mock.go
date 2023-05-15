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
	JSONPath  string      `json:"jsonPath,omitempty"` // path to the JSON file with endpoint
	JSON      interface{} `json:"json,omitempty"`
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
