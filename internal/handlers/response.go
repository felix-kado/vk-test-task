package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorResponse is the standard format for JSON error responses.
type errorResponse struct {
	Error string `json:"error"`
}

// respondWithError sends a JSON error response with a given status code and message.
func respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(errorResponse{Error: message}); err != nil {
		slog.Error("failed to encode JSON response", slog.String("error", err.Error()))
	}
}
