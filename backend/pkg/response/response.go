package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type Problem struct {
	Error string `json:"error"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to encode JSON response", "error", err)
	}
}

func Success(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, data)
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func Error(w http.ResponseWriter, statusCode int, error string) {
	if statusCode == http.StatusInternalServerError {
		slog.Error("Something went wrong", "error", error)
	} else {
		slog.Warn("A bad request was made", "error", error)
	}

	JSON(w, statusCode, Problem{
		Error: error,
	})
}

func BadRequest(w http.ResponseWriter, error string) {
	Error(w, http.StatusBadRequest, error)
}

func InternalError(w http.ResponseWriter, error string) {
	Error(w, http.StatusInternalServerError, error)
}

func NotFound(w http.ResponseWriter, error string) {
	Error(w, http.StatusNotFound, error)
}

func Unauthorized(w http.ResponseWriter, error string) {
	Error(w, http.StatusUnauthorized, error)
}
