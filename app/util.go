package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type responseWriterWrapper struct {
	http.ResponseWriter

	statusCode int
}

func (wrp *responseWriterWrapper) WriteHeader(code int) {
	wrp.ResponseWriter.WriteHeader(code)
	wrp.statusCode = code
}

func readRequestJSON(_ context.Context, req *http.Request, object any) *appError {
	err := json.NewDecoder(req.Body).Decode(object)
	if err != nil {
		return &appError{
			Error:   err,
			Message: fmt.Sprintf("wrong request body: %v", err),
			Code:    http.StatusBadRequest,
		}
	}

	return nil
}
