package app

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/smeshkov/gomock/config"
)

var (
	errTraverseNotObject = errors.New("value is not an object in JSON traversal")
	errTraverseNotFound  = errors.New("attribute not found in JSON traversal")
	errTraverseNotString = errors.New("value is not a string in JSON traversal")
	errTraverseNoParts   = errors.New("no parts in JSON path")
)

// GET /healthcheck.
func healthcheckHandler(writer http.ResponseWriter, _ *http.Request) *appError {
	return writeResponse(writer, map[string]any{
		"status": "OK",
	})
}

// GET /version.
func versionHandler(version string) func(http.ResponseWriter, *http.Request) *appError {
	return func(writer http.ResponseWriter, _ *http.Request) *appError {
		return writeResponse(writer, map[string]any{
			"version": version,
		})
	}
}

func apiHandler(log *slog.Logger, endpoint *config.Endpoint, status int,
	jsonData []byte, proxy *Proxy, database *store) func(http.ResponseWriter, *http.Request) *appError {
	errCnt, errCodes := setupFails(endpoint)

	var ops uint64

	return func(writer http.ResponseWriter, req *http.Request) *appError {
		log.Debug("handling request", "uri", req.RequestURI)

		if endpoint.Delay > 0 {
			time.Sleep(time.Duration(endpoint.Delay) * time.Millisecond)
		}

		if appErr := handleErrorSimulation(endpoint, &ops, errCnt, errCodes, log); appErr != nil {
			return appErr
		}

		// Proxy request to the provided URL.
		if endpoint.Proxy != "" {
			proxy.ServeHTTP(writer, req)

			return nil
		}

		writer.WriteHeader(status)

		return handleResponse(log, endpoint, jsonData, database, writer, req)
	}
}

func handleErrorSimulation(endpoint *config.Endpoint, ops *uint64,
	errCnt uint64, errCodes []int, log *slog.Logger) *appError {
	if endpoint.Errors == nil {
		return nil
	}

	newOps := atomic.AddUint64(ops, 1)
	if newOps != errCnt {
		return nil
	}

	atomic.StoreUint64(ops, 0)

	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(errCodes))))
	code := errCodes[idx.Int64()]

	return &appError{
		Code:    code,
		Message: "failed with predefined error",
		Log:     log,
	}
}

func handleResponse(log *slog.Logger, endpoint *config.Endpoint, jsonData []byte,
	database *store, writer http.ResponseWriter, req *http.Request) *appError {
	// Serve static JSON file from JSONPath if set.
	if endpoint.JSONPath != "" {
		_, err := writer.Write(jsonData)
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
		log.Debug("returning JSON object", "object", fmt.Sprintf("%#v", endpoint.JSON))

		return writeResponse(writer, endpoint.JSON)
	}

	// Dynamic read/write operation.
	if endpoint.Dynamic != nil {
		return handleDynamic(log, endpoint, database, writer, req)
	}

	return nil
}

func handleDynamic(log *slog.Logger, endpoint *config.Endpoint,
	database *store, writer http.ResponseWriter, req *http.Request) *appError {
	if endpoint.Dynamic.Write != nil {
		return handleDynamicWrite(log, endpoint, database, req)
	}

	if endpoint.Dynamic.Read != nil {
		return handleDynamicRead(log, endpoint, database, writer, req)
	}

	return nil
}

func handleDynamicWrite(log *slog.Logger, endpoint *config.Endpoint,
	database *store, req *http.Request) *appError {
	input := map[string]any{}

	appErr := readRequestJSON(req.Context(), req, &input)
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

	log.Debug("writing dynamic entry", "name", endpoint.Dynamic.Write.JSON.Name, "key", key)
	database.Write(endpoint.Dynamic.Write.JSON.Name, key, value)

	return nil
}

func handleDynamicRead(log *slog.Logger, endpoint *config.Endpoint,
	database *store, writer http.ResponseWriter, req *http.Request) *appError {
	var (
		key   string
		value any
		found bool
	)

	if endpoint.Dynamic.Read.JSON.KeyParam == "" {
		value, found = database.ReadAll(endpoint.Dynamic.Read.JSON.Name)
	} else {
		key = chi.URLParam(req, endpoint.Dynamic.Read.JSON.KeyParam)
		value, found = database.Read(endpoint.Dynamic.Read.JSON.Name, key)
	}

	if !found {
		return &appError{
			Message: fmt.Sprintf("value not found for key [%s]", key),
			Log:     log,
		}
	}

	log.Debug("reading dynamic entry", "name", endpoint.Dynamic.Read.JSON.Name, "key", key)

	return writeResponse(writer, value)
}

func setupFails(endpoint *config.Endpoint) (uint64, []int) {
	if endpoint.Errors == nil {
		return 0, nil
	}

	var errCodes []int

	if len(endpoint.Errors.Statuses) > 0 {
		errCodes = endpoint.Errors.Statuses
	} else {
		errCodes = []int{http.StatusInternalServerError}
	}

	errCnt := uint64(1.0 / endpoint.Errors.Sample)

	slog.Debug("every Nth request will fail",
		"endpoint", endpoint.Path,
		"every_nth_err", errCnt,
		"errCodes", fmt.Sprintf("%v", errCodes))

	return errCnt, errCodes
}

func setupAPI(cfg *config.Config, mockPath string, mck *config.Mock, router *chi.Mux) {
	database := newStore()

	for _, endpoint := range mck.Endpoints {
		status := endpoint.Status
		if status <= 0 {
			status = http.StatusOK
		}

		logger := slog.Default().With(
			"endpoint", endpoint.Path,
			"methods", fmt.Sprintf("%v", endpoint.Methods),
		)

		route := resolveRoute(endpoint.Path)

		logger = logger.With("route", route)
		logger.Info("setting up endpoint")

		configureRoute(router, route, cfg, mockPath, endpoint, logger, status, database)
	}
}

func resolveRoute(endpointPath string) string {
	switch {
	case endpointPath == "/" || endpointPath == "*" || endpointPath == "":
		return "/"
	case strings.HasSuffix(endpointPath, "/*"):
		return path.Dir(endpointPath)
	case strings.Contains(endpointPath, "*"):
		idx := 0
		newPath := endpointPath

		for strings.Contains(newPath, "*") {
			newPath = strings.Replace(newPath, "*", fmt.Sprintf("{subpath-%d}", idx), 1)
			idx++
		}

		return newPath
	default:
		return endpointPath
	}
}

func configureRoute(router *chi.Mux, route string, cfg *config.Config, mockPath string,
	endpoint *config.Endpoint, logger *slog.Logger, status int, database *store) {
	router.Route(route, func(subrouter chi.Router) {
		if len(endpoint.AllowCors) > 0 {
			subrouter.Use(NewCORS(endpoint.AllowCors...).Middleware)
		}

		var err error

		var jsonData []byte

		if endpoint.JSONPath != "" {
			jsonData, err = readJSON(mockPath, endpoint.JSONPath)
			if err != nil {
				logger.Error(fmt.Sprintf("error in reading JSON from the path [%s] for path [%s]",
					endpoint.JSONPath, endpoint.Path))

				return
			}
		}

		var proxy *Proxy

		if endpoint.Proxy != "" {
			proxy, err = newProxy(cfg.Server.Addr, endpoint.Proxy, logger)
			if err != nil {
				logger.Error(fmt.Sprintf("error in creating a proxy for path [%s]: %v",
					endpoint.Path, err))

				return
			}
		}

		if endpoint.Static != "" {
			subrouter.HandleFunc("/*", http.FileServer(http.Dir(endpoint.Static)).ServeHTTP)

			return
		}

		for _, method := range endpoint.Methods {
			subrouter.Method(method, "/*", appHandler(apiHandler(logger, endpoint, status, jsonData, proxy, database)))
		}
	})
}

func readJSON(mockPath, jsonPath string) ([]byte, error) {
	filePath := filepath.Join(mockPath, jsonPath)

	data, err := os.ReadFile(filepath.Clean(filePath))
	if err != nil {
		return nil, fmt.Errorf("reading JSON file %s: %w", filePath, err)
	}

	return data, nil
}

func findKeyInJSON(jsonPath string, obj map[string]any) (string, error) {
	parts := strings.Split(jsonPath, "/")

	var current any = obj

	for idx, part := range parts {
		mapVal, isMap := current.(map[string]any)
		if !isMap {
			return "", fmt.Errorf("%w: [%s] for the key [%s]", errTraverseNotObject, part, jsonPath)
		}

		val, exists := mapVal[part]
		if !exists {
			return "", fmt.Errorf("%w: [%s] for the key [%s]", errTraverseNotFound, part, jsonPath)
		}

		if idx < len(parts)-1 {
			current = val
		} else {
			strVal, isStr := val.(string)
			if !isStr {
				return "", fmt.Errorf("%w: [%s] for the key [%s]", errTraverseNotString, part, jsonPath)
			}

			return strVal, nil
		}
	}

	return "", fmt.Errorf("%w: [%s]", errTraverseNoParts, jsonPath)
}

func findValueInJSON(jsonPath string, obj map[string]any) (any, error) {
	parts := strings.Split(jsonPath, "/")

	var current any = obj

	for idx, part := range parts {
		if part == "" || part == "." {
			continue
		}

		mapVal, isMap := current.(map[string]any)
		if !isMap {
			return "", fmt.Errorf("%w: [%s] for the value [%s]", errTraverseNotObject, part, jsonPath)
		}

		val, exists := mapVal[part]
		if !exists {
			return "", fmt.Errorf("%w: [%s] for the value [%s]", errTraverseNotFound, part, jsonPath)
		}

		if idx < len(parts)-1 {
			current = val
		} else {
			return val, nil
		}
	}

	return current, nil
}
