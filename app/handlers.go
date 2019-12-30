package app

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sync/atomic"
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
	var ops uint64
	var errCnt uint64
	var errCodes []int
	if endpoint.Errors != nil {
		if len(endpoint.Errors.Statuses) > 0 {
			errCodes = endpoint.Errors.Statuses
		} else {
			errCodes = []int{http.StatusInternalServerError}
		}
		errCnt = uint64(1.0 / endpoint.Errors.Sample)
		c.Log.Debug("%s - every %d request will fail with either of %v HTTP codes", endpoint.Path, errCnt, errCodes)
	}

	return func(w http.ResponseWriter, r *http.Request) *appError {

		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		if endpoint.Errors != nil {
			atomic.AddUint64(&ops, 1)
			if ops == errCnt {
				atomic.StoreUint64(&ops, 0)
				code := errCodes[rand.Intn(len(errCodes))]
				c.Log.Debug("%s - HTTP %d", endpoint.Path, code)
				return &appError{
					Code: code,
				}

			}
		}

		// Proxy request to the provide URL.
		if endpoint.URL != "" {
			c.Log.Debug("%s - proxying to %s", endpoint.Path, endpoint.URL)
			u, err := url.Parse(endpoint.URL)
			if err != nil {
				return appErrorf(err, "error in parsing URL %s", endpoint.URL)
			}
			resp, err := client.proxy(r, u)
			if err != nil {
				return appErrorf(err, "error in proxying to %s", endpoint.URL)
			}
			for k, vs := range resp.Header {
				for _, v := range vs {
					w.Header().Add(k, v)
				}
			}
			buf := new(bytes.Buffer)
			if resp.Body != nil {
				_, err = buf.ReadFrom(resp.Body)
				if err != nil {
					return appErrorf(err, "error in reading response from %s", endpoint.URL)
				}
				defer resp.Body.Close()
				_, err = w.Write(buf.Bytes())
				if err != nil {
					return appErrorf(err, "error in writing response from %s", endpoint.URL)
				}
			}
			return nil
		}

		w.WriteHeader(status)

		// Serve static JSON file from JSONPath if set.
		if endpoint.JSONPath != "" {
			c.Log.Debug("%s - replying with %s", endpoint.Path, endpoint.JSONPath)
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
		c.Log.Debug("%s - replying with %#v", endpoint.Path, endpoint.JSON)
		writeResponse(w, endpoint.JSON)
		return nil
	}
}

func setupAPI(mck *c.Mock, router *mux.Router, client *client) {
	for _, e := range mck.Endpoints {

		c.Log.Info("setting up %s", e.Path)

		var r *mux.Route
		if e.Path == "/" {
			r = router.PathPrefix(e.Path)
		} else {
			r = router.Path(e.Path)
		}

		r = r.Methods(e.Methods...)

		status := e.Status
		if status <= 0 {
			status = http.StatusOK
		}

		r.Handler(appHandler(apiHandler(e, status, client)))
	}
}
