package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// readRequestJSON ...
func readRequestJSON(c context.Context, r *http.Request, object interface{}) *appError {
	err := json.NewDecoder(r.Body).Decode(object)
	if err != nil {
		return &appError{
			Error:   err,
			Message: fmt.Sprintf("wrong request body: %v", err),
			Code:    http.StatusBadRequest,
		}
	}
	return nil
}
