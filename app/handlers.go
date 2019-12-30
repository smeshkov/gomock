package app

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	c "github.com/smeshkov/gomock/config"
)

var zeroDuration = time.Duration(0)

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

func apiHandler(endpoint *c.Endpoint, status int, client *client) func(rw http.ResponseWriter, req *http.Request) *appError {
	return func(w http.ResponseWriter, r *http.Request) *appError {

		c.Log.Debug("accessed %s", endpoint.Path)

		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		// Proxy request to the provide URL.
		if endpoint.URL != "" {
			c.Log.Debug("proxying to %s", endpoint.URL)
			u, err := url.Parse(endpoint.URL)
			if err != nil {
				return appErrorf(err, "error in parsing URL %s", endpoint.URL)
			}
			resp, err := client.proxy(r, u)
			if err != nil {
				return appErrorf(err, "error in proxying to %s", endpoint.URL)
			}
			defer resp.Body.Close()
			w.WriteHeader(resp.StatusCode)
			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(resp.Body)
			if err != nil {
				return appErrorf(err, "error in reading response from %s", endpoint.URL)
			}
			_, err = w.Write(buf.Bytes())
			if err != nil {
				return appErrorf(err, "error in writing response from %s", endpoint.URL)
			}
			return nil
		}

		w.WriteHeader(status)

		// Serve static JSON file from JSONPath if set.
		if endpoint.JSONPath != "" {
			c.Log.Debug("replying from %s", endpoint.JSONPath)
			data, err := ioutil.ReadFile(endpoint.JSONPath)
			if err != nil {
				return appErrorf(err, "error in reading from %s", endpoint.JSONPath)
			}
			_, err = w.Write(data)
			if err != nil {
				return appErrorf(err, "error in writing data from %s", endpoint.JSONPath)
			}
			return nil
		}

		// Serve JSON from API configuration instead.
		c.Log.Debug("replying with %#v", endpoint.JSON)
		writeResponse(w, endpoint.JSON)
		return nil
	}
}

func setupAPI(mck *c.Mock, router *mux.Router, client *client) {
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
			Handler(appHandler(apiHandler(e, status, client)))
	}
}
