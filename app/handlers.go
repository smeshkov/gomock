package app

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"
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

func apiHandler(log *zap.Logger, endpoint *config.Endpoint, status int,
	jsonData []byte, proxy *Proxy, db *store) func(http.ResponseWriter, *http.Request) *appError {
	errCnt, errCodes := setupFails(endpoint)
	var ops uint64

	return func(w http.ResponseWriter, r *http.Request) *appError {

		log.Sugar().Debugf("handling [%s]", r.RequestURI)

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
			proxy.ServeHTTP(w, r)
			return nil
		}

		w.WriteHeader(status)

		// Serve static JSON file from JSONPath if set.
		if endpoint.JSONPath != "" {
			_, err := w.Write(jsonData)
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
		if endpoint.JSON != nil {
			log.Debug("returning JSON object", zap.String("object", fmt.Sprintf("%#v", endpoint.JSON)))
			return writeResponse(w, endpoint.JSON)
		}

		// Dynamic read/write operation
		if endpoint.Dynamic != nil {
			if endpoint.Dynamic.Write != nil {
				input := map[string]interface{}{}
				appErr := readRequestJSON(r.Context(), r, &input)
				if appErr != nil {
					return appErr
				}
				key, err := findKeyInJSON(endpoint.Dynamic.Write.JSON.Key, input)
				if err != nil {
					return &appError{
						Error:   err,
						Message: "error in finding the key",
						Log:     log,
					}
				}
				value, err := findValueInJSON(endpoint.Dynamic.Write.JSON.Value, input)
				if err != nil {
					return &appError{
						Error:   err,
						Message: "error in finding the value",
						Log:     log,
					}
				}
				log.Sugar().Debugf("writing [%s] under the key [%s]", endpoint.Dynamic.Write.JSON.Name, key)
				db.Write(endpoint.Dynamic.Write.JSON.Name, key, value)
			} else if endpoint.Dynamic.Read != nil {
				var key string
				var value interface{}
				var ok bool
				if endpoint.Dynamic.Read.JSON.KeyParam == "" {
					value, ok = db.ReadAll(endpoint.Dynamic.Read.JSON.Name)
				} else {
					key = chi.URLParam(r, endpoint.Dynamic.Read.JSON.KeyParam)
					value, ok = db.Read(endpoint.Dynamic.Read.JSON.Name, key)
				}
				if !ok {
					return &appError{
						Message: fmt.Sprintf("value not found for key [%s]", key),
						Log:     log,
					}
				}
				log.Sugar().Debugf("reading [%s] under the key [%s]", endpoint.Dynamic.Read.JSON.Name, key)
				return writeResponse(w, value)
			}
		}

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
			zap.String("endpoint", endpoint.Path),
			zap.Uint64("every_nth_err", errCnt),
			zap.String("errCodes", fmt.Sprintf("%v", errCodes)))
	}
	return
}

func setupAPI(cfg *config.Config, mockPath string, mck *config.Mock, router *chi.Mux) {
	db := newStore()
	for _, e := range mck.Endpoints {

		status := e.Status
		if status <= 0 {
			status = http.StatusOK
		}

		l := zap.L().With(
			zap.String("endpoint", e.Path),
			zap.String("methods", fmt.Sprintf("%v", e.Methods)),
		)

		if l.Core().Enabled(zap.DebugLevel) {
			l = l.With(
				zap.Int("status", status),
				zap.Bool("json", e.JSON != nil),
				zap.Bool("dynamic", e.Dynamic != nil),
				zap.String("jsonPath", e.JSONPath),
				zap.String("proxy", e.Proxy),
				zap.String("static", e.Static),
			)
		}

		var route string
		if e.Path == "/" || e.Path == "*" || e.Path == "" {
			route = "/"
		} else if strings.HasSuffix(e.Path, "/*") {
			route = path.Dir(e.Path)
		} else if strings.Contains(e.Path, "*") {
			var i int
			newPath := e.Path
			for strings.Contains(newPath, "*") {
				newPath = strings.Replace(newPath, "*", fmt.Sprintf("{subpath-%d}", i), 1)
				i++
			}
			route = newPath
		} else {
			route = e.Path
		}

		l = l.With(zap.String("route", route))
		l.Info("setting up endpoint")

		router.Route(route, func(r chi.Router) {
			if len(e.AllowCors) > 0 {
				r.Use(NewCORS(e.AllowCors...).Middleware)
			}

			var err error

			var jsonData []byte
			if e.JSONPath != "" {
				jsonData, err = readJSON(mockPath, e.JSONPath)
				if err != nil {
					l.Sugar().
						Errorf("error in reading JSON from the path [%s] for path [%s]", e.JSONPath, e.Path)
					return
				}
			}

			var proxy *Proxy
			if e.Proxy != "" {
				proxy, err = newProxy(cfg.Server.Addr, e.Proxy, l)
				if err != nil {
					l.Sugar().
						Errorf("error in creating a proxy for path [%s]: %v", e.Path, err)
					return
				}
			}

			if e.Static != "" {
				r.HandleFunc("/*", http.FileServer(http.Dir(e.Static)).ServeHTTP)
				return
			}

			for _, m := range e.Methods {
				r.Method(m, "/*", appHandler(apiHandler(l, e, status, jsonData, proxy, db)))
			}
		})

	}
}

func readJSON(mockPath, jsonPath string) ([]byte, error) {
	p := filepath.Join(mockPath, jsonPath)
	return os.ReadFile(p)
}

func findKeyInJSON(path string, obj map[string]interface{}) (string, error) {
	parts := strings.Split(path, "/")
	v := obj
	for i, p := range parts {
		v, ok := v[p]
		if !ok {
			return "", fmt.Errorf("error in traversing request JSON, attribute [%s] not found for the key [%s]", p, path)
		}
		if i < len(parts)-1 {
			v, ok = v.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("error in traversing request JSON, value of [%s] is not an object for the key [%s]", p, path)
			}
		} else {
			s, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("error in traversing request JSON, value of key [%s] is not a string for the key [%s]", p, path)
			}
			return s, nil
		}
	}
	return "", fmt.Errorf("error in traversing request JSON, no parts for the key [%s]", path)
}

func findValueInJSON(path string, obj map[string]interface{}) (interface{}, error) {
	parts := strings.Split(path, "/")
	v := obj
	for i, p := range parts {
		if p == "" || p == "." {
			continue
		}
		v, ok := v[p]
		if !ok {
			return "", fmt.Errorf("error in traversing request JSON, attribute [%s] not found for the value [%s]", p, path)
		}
		if i < len(parts)-1 {
			v, ok = v.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("error in traversing request JSON, value of [%s] is not an object for the value [%s]", p, path)
			}
		} else {
			return v, nil
		}
	}
	return v, nil
}
