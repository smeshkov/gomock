package app

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	c "github.com/smeshkov/gomock/config"
)

// GET /healthcheck
func healthcheckHandler(w http.ResponseWriter, r *http.Request) *appError {
	return writeResponse(w, map[string]interface{}{
		"status": "OK",
	})
}

// GET /version
func versionHandler(version string) func(rw http.ResponseWriter, req *http.Request) *appError {
	return func(w http.ResponseWriter, r *http.Request) *appError {
		return writeResponse(w, map[string]interface{}{
			"version": version,
		})
	}
}

func apiHandler(endpoint *c.Endpoint, status int) func(rw http.ResponseWriter, req *http.Request) *appError {
	return func(w http.ResponseWriter, r *http.Request) *appError {

		c.Log.Debug("accessed %s", endpoint.Path)

		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)

		// Serve static JSON file from JSONPath if set.
		if endpoint.JSONPath != "" {
			c.Log.Debug("replying from %s", endpoint.JSONPath)
			data, err := ioutil.ReadFile(endpoint.JSONPath)
			if err != nil {
				return appErrorf(err, "error in reading from %s", endpoint.JSONPath)
			}
			w.Write(data)
			return nil
		}

		// Serve JSON from API configuration instead.
		c.Log.Debug("replying with %#v", endpoint.JSON)
		writeResponse(w, endpoint.JSON)
		return nil
	}
}

func setupAPI(mck *c.Mock, router *mux.Router) error {
	if mck == nil {
		return errors.New("provided API is nil")
	}

	for _, e := range mck.Endpoints {

		c.Log.Info("setting up %s", e.Path)

		method := e.Method
		if method == "" {
			method = http.MethodGet
		}

		status := e.Status
		if status <= 0 {
			status = http.StatusOK
		}

		router.
			Methods(method).
			Path(e.Path).
			Handler(appHandler(apiHandler(e, status)))
	}

	return nil
}
