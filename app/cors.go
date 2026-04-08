package app

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/handlers"
)

// CORS provides cross-origin resource sharing middleware.
type CORS struct {
	allowedOrigins []string
	handler        func(http.Handler) http.Handler
	IsAllowed      func(string) bool
}

// NewCORS creates new CORS.
func NewCORS(allowedOrigins ...string) *CORS {
	isAllowed := func(origin string) bool {
		for _, allowed := range allowedOrigins {
			if allowed == "*" || origin == allowed || strings.HasSuffix(origin, allowed) {
				return true
			}
		}

		slog.Debug("CORS - not allowed origin: " + origin)

		return false
	}

	return &CORS{
		allowedOrigins: allowedOrigins,
		handler: handlers.CORS(
			handlers.AllowedOriginValidator(isAllowed),
			handlers.AllowedOrigins(allowedOrigins),
			handlers.AllowedMethods([]string{
				"OPTIONS",
				"HEAD",
				"GET",
				"POST",
				"PUT",
				"DELETE",
			}),
			handlers.AllowedHeaders([]string{
				"Accept",
				"X-Requested-With",
				"Authorization",
				"Content-Type",
				"Access-Control-Max-Age",
			}),
			handlers.AllowCredentials(),
		),
		IsAllowed: isAllowed,
	}
}

// Middleware returns the CORS middleware handler.
func (c *CORS) Middleware(next http.Handler) http.Handler {
	return c.handler(next)
}
