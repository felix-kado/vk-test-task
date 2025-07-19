package http

import (
	"log/slog"
	"net/http"

	"example.com/market/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// NewRouter creates a new chi router and sets up the routes and middlewares.
func NewRouter(log *slog.Logger, authHandler *AuthHandler, adsHandler *AdsHandler, authService middleware.TokenParser) *chi.Mux {
	r := chi.NewRouter()

	// Base middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestLogger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.URLFormat)

	// Health check
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Public routes
	r.Post("/v1/register", authHandler.Register)
	r.Post("/v1/login", authHandler.Login)

	// Public ad routes
	r.Get("/v1/ads", adsHandler.ListAds)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthCtx(authService))
		r.Post("/v1/ads", adsHandler.CreateAd)
	})

	return r
}
