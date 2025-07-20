package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"example.com/market/internal/domain"
	"example.com/market/internal/services"
)

// AuthService defines the interface for authentication-related operations.
type AuthService interface {
	Register(ctx context.Context, login, password string) (string, *domain.User, error)
	Login(ctx context.Context, login, password string) (string, error)
}

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	service AuthService
	log     *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service AuthService, log *slog.Logger) *AuthHandler {
	return &AuthHandler{service: service, log: log}
}

// RegistrationRequest defines the structure for a user registration request.
type RegistrationRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user and returns their ID.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   input body RegistrationRequest true "Registration Info"
// @Success 201 {object} map[string]int64
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /register [post]
// Register handles user registration requests.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	_, user, err := h.service.Register(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrUserExists) {
			h.log.Info("user already exists", slog.String("login", req.Login))
			respondWithError(w, http.StatusConflict, "user with this login already exists")
			return
		}
		h.log.Error("failed to register user", slog.String("error", err.Error()))
		respondWithError(w, http.StatusInternalServerError, "could not register user")
		return
	}

	resp := map[string]int64{"user_id": user.ID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// LoginRequest defines the structure for a user login request.
type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Login godoc
// @Summary Log in a user
// @Description Authenticates a user and returns a JWT token.
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   input body LoginRequest true "Login Credentials"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /login [post]
// Login handles user login requests.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.service.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, services.ErrInvalidCredentials) {
			respondWithError(w, http.StatusUnauthorized, "invalid login or password")
			return
		}
		h.log.Error("failed to login", slog.String("error", err.Error()))
		respondWithError(w, http.StatusInternalServerError, "an internal error occurred")
		return
	}

	resp := map[string]string{"token": token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
