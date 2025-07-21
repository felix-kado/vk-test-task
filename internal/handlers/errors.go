package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/felix-kado/vk-test-task/internal/services"
)

// handleServiceError maps service layer errors to HTTP responses.
func handleServiceError(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	switch {
	case errors.Is(err, services.ErrInvalidInput):
		respondWithError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrConflict):
		respondWithError(w, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrAdNotFound):
		respondWithError(w, http.StatusNotFound, "ad not found")
	case errors.Is(err, services.ErrUserNotFound):
		respondWithError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, services.ErrUnauthorized):
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, services.ErrForbidden):
		respondWithError(w, http.StatusForbidden, "forbidden")
	default:
		// For unhandled errors, log them and return a generic 500 response.
		log.Error("internal server error", slog.String("path", r.URL.Path), slog.String("error", err.Error()))
		respondWithError(w, http.StatusInternalServerError, "internal server error")
	}
}
