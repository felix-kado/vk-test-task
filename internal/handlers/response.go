package handlers

import (
	"encoding/json"
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
	json.NewEncoder(w).Encode(errorResponse{Error: message})
}
