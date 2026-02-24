package wrapper

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"

	"github.com/kuromii5/chat-bot-chat-service/pkg/validator"
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
		"service":    "chat-service",
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

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		logrus.WithError(err).Error("json encode failed")
	}
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Success(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		logrus.WithError(err).Error("encode json: %w", err)
	}
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
