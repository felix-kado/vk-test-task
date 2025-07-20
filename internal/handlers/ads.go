package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"example.com/market/internal/domain"
	"example.com/market/internal/middleware"
	"example.com/market/internal/storage"
)

// AdsService defines the interface for ad-related operations.
type AdsService interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
}

// AdsHandler handles HTTP requests for ads.
type AdsHandler struct {
	service AdsService
	log     *slog.Logger
}

// NewAdsHandler creates a new AdsHandler.
func NewAdsHandler(service AdsService, log *slog.Logger) *AdsHandler {
	return &AdsHandler{service: service, log: log}
}

// AdRequest defines the structure for an ad creation request.
type AdRequest struct {
	Title    string `json:"title"`
	Text     string `json:"text"`
	ImageURL string `json:"image_url,omitempty"`
	Price    int64  `json:"price,omitempty"`
}

// CreateAd godoc
// @Summary Create a new ad
// @Security ApiKeyAuth
// @Description Creates a new ad for the authenticated user.
// @Tags ads
// @Accept  json
// @Produce  json
// @Param   input body AdRequest true "Ad Info"
// @Success 201 {object} map[string]int64
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ads [post]
// CreateAd handles ad creation requests.
func (h *AdsHandler) CreateAd(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int64)
	if !ok {
		h.log.Error("failed to get user ID from context")
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	var req AdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate ad request
	if err := ValidateAdRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ad := &domain.Ad{
		UserID:   userID,
		Title:    req.Title,
		Text:     req.Text,
		ImageURL: req.ImageURL,
		Price:    req.Price,
	}

	adID, err := h.service.CreateAd(r.Context(), ad)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "user not found", http.StatusBadRequest)
			return
		}
		h.log.Error("failed to create ad", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := map[string]int64{"id": adID}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

// ListAds godoc
// @Summary List ads
// @Description Returns a list of ads, with optional sorting.
// @Tags ads
// @Produce  json
// @Param   sort_by query string false "Sort by field (price or created_at)" Enums(price, created_at)
// @Param   order query string false "Sort order (asc or desc)" Enums(asc, desc)
// @Success 200 {array} domain.Ad
// @Failure 500 {object} map[string]string
// @Router /ads [get]
// ListAds handles requests to list ads.
func (h *AdsHandler) ListAds(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at" // default
	}

	order := r.URL.Query().Get("order")
	if order == "" {
		order = "desc" // default
	}

	// Validate query parameters
	if err := ValidateListParams(sortBy, order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ads, err := h.service.ListAds(r.Context(), sortBy, order)
	if err != nil {
		h.log.Error("failed to list ads", slog.String("error", err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		h.log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func ValidateAdRequest(req *AdRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if req.Text == "" {
		return errors.New("text is required")
	}
	if req.Price < 0 {
		return errors.New("price must be non-negative")
	}
	return nil
}

func ValidateListParams(sortBy, order string) error {
	if sortBy != "price" && sortBy != "created_at" {
		return errors.New("invalid sort_by parameter")
	}
	if order != "asc" && order != "desc" {
		return errors.New("invalid order parameter")
	}
	return nil
}
