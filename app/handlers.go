package app

import (
	"errors"
	"net/http"

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

func setupAPI(api *c.API, router *mux.Router) error {
	if api == nil {
		return errors.New("provided API is nil")
	}

	for _, e := range api.Endpoints {
		router.
			Methods(e.Method).
			Path(e.Path).
			Handler(appHandler(func(w http.ResponseWriter, r *http.Request) *appError {

				w.Header().Set("Content-Type", "application/json")

				// Serve static JSON file from JSONPath if set
				if e.JSONPath != "" {
					http.ServeFile(w, r, e.JSONPath)
					return nil
				}

				// Serve JSON from API configuration instead
				w.Write([]byte(e.JSON))
				return nil
			}))
	}

	return nil
}
