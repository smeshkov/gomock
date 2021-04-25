package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/smeshkov/gomock/config"
)

// var zeroDuration = time.Duration(0)

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

func apiHandler(log *zap.Logger, endpoint *config.Endpoint, status int, client *client,
	mockPath string) func(http.ResponseWriter, *http.Request) *appError {
	errCnt, errCodes := setupFails(endpoint)
	var ops uint64

	return func(w http.ResponseWriter, r *http.Request) *appError {

		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		if endpoint.Errors != nil {
			atomic.AddUint64(&ops, 1)
			if ops == errCnt {
				atomic.StoreUint64(&ops, 0)
				code := errCodes[rand.Intn(len(errCodes))]
				return &appError{
					Code:    code,
					Message: "failed with predefined error",
					Log:     log,
				}

			}
		}

		// Proxy request to the provided URL.
		if endpoint.Proxy != "" {
			log.Debug("proxying call", zap.String("proxy_to", endpoint.Proxy))
			u, err := url.Parse(endpoint.Proxy)
			if err != nil {
				return &appError{
					Error:   err,
					Message: "error in parsing proxy to URL",
					Log:     log,
				}
			}
			resp, err := client.proxy(r, u)
			if err != nil {
				return &appError{
					Error:   err,
					Message: "error in proxying to URL",
					Log:     log,
				}
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
					return &appError{
						Error:   err,
						Message: "error in reading response from proxyed URL",
						Log:     log,
					}
				}
				defer resp.Body.Close()
				_, err = w.Write(buf.Bytes())
				if err != nil {
					return &appError{
						Error:   err,
						Message: "error in writing response to client",
						Log:     log,
					}
				}
			}
			return nil
		}

		w.WriteHeader(status)

		// Serve static JSON file from JSONPath if set.
		if endpoint.JSONPath != "" {
			p := filepath.Join(mockPath, endpoint.JSONPath)
			log.Debug("returning contents of JSON path", zap.String("json_path", endpoint.JSONPath))
			data, err := ioutil.ReadFile(p)
			if err != nil {
				return &appError{
					Error:   err,
					Message: fmt.Sprintf("error in reading from JSON path: [%s]", p),
					Log:     log,
				}
			}
			_, err = w.Write(data)
			if err != nil {
				return &appError{
					Error:   err,
					Message: "error in writing data from JSON path to client",
					Log:     log,
				}
			}
			return nil
		}

		// Serve JSON from API configuration instead.
		log.Debug("returning object", zap.String("object", fmt.Sprintf("%#v", endpoint.JSON)))
		writeResponse(w, endpoint.JSON)
		return nil
	}
}

func setupFails(endpoint *config.Endpoint) (errCnt uint64, errCodes []int) {
	if endpoint.Errors != nil {
		if len(endpoint.Errors.Statuses) > 0 {
			errCodes = endpoint.Errors.Statuses
		} else {
			errCodes = []int{http.StatusInternalServerError}
		}
		errCnt = uint64(1.0 / endpoint.Errors.Sample)
		zap.L().Debug("every Nth request will fail",
			zap.String("path", endpoint.Path),
			zap.Uint64("every_nth_err", errCnt),
			zap.String("errCodes", fmt.Sprintf("%v", errCodes)))
	}
	return
}

func setupAPI(mockPath string, mck *config.Mock, router *mux.Router, client *client) {
	for _, e := range mck.Endpoints {

		status := e.Status
		if status <= 0 {
			status = http.StatusOK
		}

		l := zap.L().With(
			zap.String("path", e.Path),
			zap.String("methods", fmt.Sprintf("%v", e.Methods)),
			zap.Int("status", status),
		)

		var r *mux.Router
		if e.Path == "/" {
			r = router.PathPrefix(e.Path).Subrouter()
		} else if strings.HasSuffix(e.Path, "*") {
			r = router.PathPrefix(path.Dir(e.Path)).Subrouter()
		} else if e.Path == "*" || e.Path == "" {
			r = router
		} else {
			r = router.Path(e.Path).Subrouter()
		}

		if len(e.AllowCors) > 0 {
			r.Use(NewCORS(e.AllowCors...).Middleware)
		}

		l.Info("setting up path")

		r.Methods(e.Methods...).Handler(appHandler(apiHandler(l, e, status, client, mockPath)))
	}
}
