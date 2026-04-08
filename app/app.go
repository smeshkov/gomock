// Package app contains the HTTP layer for the mock server.
package app

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/smeshkov/gomock/config"
)

// RegisterHandlers registers all handlers of the application.
func RegisterHandlers(version, mockPath string, cfg *config.Config, mck *config.Mock) http.Handler {
	router := chi.NewRouter()

	// Shows if app is healthy
	router.Method(http.MethodGet, "/healthcheck", appHandler(healthcheckHandler))

	// Shows current version of the App
	router.Method(http.MethodGet, "/version", appHandler(versionHandler(version)))

	setupAPI(cfg, mockPath, mck, router)

	return router
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
	Log     *slog.Logger
}

func (fn appHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	appErr := fn(writer, req)
	if appErr != nil {
		logger := appErr.Log
		if logger == nil {
			logger = slog.Default()
		}

		logger.Error(fmt.Sprintf("handler error: status code: %d, message: %s, underlying err: %#v",
			appErr.Code, appErr.Message, appErr.Error))

		http.Error(writer, appErr.Message, appErr.Code)
	}
}

// writeResponse writes response to provided ResponseWriter in JSON format.
func writeResponse(writer http.ResponseWriter, response any) *appError {
	writer.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(writer).Encode(response)
	if err != nil {
		return &appError{
			Error:   err,
			Message: fmt.Sprintf("error in response write: %v", err),
			Code:    http.StatusInternalServerError,
		}
	}

	return nil
}
