package wrapper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
	"github.com/sirupsen/logrus"
)

type errorMapping struct {
	status  int
	message string
	details any
}

func WrapError(w http.ResponseWriter, r *http.Request, err error) {
	mapping := getErrorMapping(err)

	requestID := middleware.GetReqID(r.Context())

	logrus.WithFields(logrus.Fields{
		"request_id": requestID,
		"status":     mapping.status,
		"error":      err.Error(),
		"method":     r.Method,
		"path":       r.URL.Path,
	}).Log(getLogLevel(mapping.status), "request_error")

	response := map[string]any{"error": mapping.message}
	if mapping.details != nil {
		response["details"] = mapping.details
	}

	JSON(w, mapping.status, response)
}

func JSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func NoContent(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func Success(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getErrorMapping(err error) errorMapping {
	for domErr, resp := range errorRegistry {
		if errors.Is(err, domErr) {
			return errorMapping{status: resp.Status, message: resp.Message}
		}
	}

	var (
		vErr         validator.ValidationError
		syntaxErr    *json.SyntaxError
		unmarshalErr *json.UnmarshalTypeError
	)

	switch {
	case errors.As(err, &vErr):
		return errorMapping{http.StatusBadRequest, "Validation failed", vErr.Fields}

	case errors.As(err, &syntaxErr), errors.Is(err, io.EOF):
		return errorMapping{status: http.StatusBadRequest, message: "Invalid JSON body"}

	case errors.As(err, &unmarshalErr):
		return errorMapping{status: http.StatusBadRequest, message: "Wrong data types in JSON"}
	}

	return errorMapping{status: http.StatusInternalServerError, message: "Internal server error"}
}

func getLogLevel(status int) logrus.Level {
	if status >= 500 {
		return logrus.ErrorLevel
	}
	return logrus.WarnLevel
}
