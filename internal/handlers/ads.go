package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/market/internal/domain"
	"example.com/market/internal/dto"
	"example.com/market/internal/middleware"
)

// AdsService defines the interface for ad-related operations.
type AdsService interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error)
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
		h.log.Error("unauthorized: missing user ID in context")
		respondWithError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req AdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
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
		handleServiceError(w, r, h.log, err)
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
// @Security ApiKeyAuth
// @Description Returns a list of ads with pagination and filtering.
// @Tags ads
// @Produce  json
// @Param   sort_by query string false "Sort by field (price or created_at)" Enums(price, created_at)
// @Param   order query string false "Sort order (asc or desc)" Enums(asc, desc)
// @Param   page query int false "Page number (1-based)"
// @Param   limit query int false "Number of items per page (max 100)"
// @Param   min_price query int false "Minimum price filter"
// @Param   max_price query int false "Maximum price filter"
// @Success 200 {array} dto.AdResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ads [get]
// ListAds handles requests to list ads.
func (h *AdsHandler) ListAds(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	params, err := h.parseListAdsParams(r)
	if err != nil {
		h.log.Warn("invalid query parameters", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	ads, err := h.service.ListAds(r.Context(), params)
	if err != nil {
		handleServiceError(w, r, h.log, err)
		return
	}

	// Get current user ID from context, if available
	rawUserID := r.Context().Value(middleware.UserIDKey)
	h.log.Debug("raw user_id from context", slog.Any("raw_user_id", rawUserID), slog.String("type", fmt.Sprintf("%T", rawUserID)))

	currentUserID, ok := rawUserID.(int64)
	if !ok {
		h.log.Debug("could not get user_id from context or user is not authenticated", slog.Any("user_id_from_context", rawUserID))
		currentUserID = 0 // Ensure it's zero if not found or wrong type
	} else {
		h.log.Debug("successfully got user_id from context", slog.Int64("user_id", currentUserID))
	}

	// Convert to DTOs
	adResponses := dto.ToAdResponseList(ads, currentUserID)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(adResponses); err != nil {
		h.log.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

// parseListAdsParams parses and validates query parameters for listing ads.
func (h *AdsHandler) parseListAdsParams(r *http.Request) (*domain.ListAdsParams, error) {
	params := &domain.ListAdsParams{}
	query := r.URL.Query()

	// Parse sorting parameters
	params.SortBy = query.Get("sort_by")
	params.Order = query.Get("order")

	// Parse pagination parameters
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: must be a number")
		}
		params.Page = page
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, fmt.Errorf("invalid limit parameter: must be a number")
		}
		params.Limit = limit
	}

	// Parse price filter parameters
	if minPriceStr := query.Get("min_price"); minPriceStr != "" {
		minPrice, err := strconv.ParseInt(minPriceStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid min_price parameter: must be a number")
		}
		params.MinPrice = &minPrice
	}

	if maxPriceStr := query.Get("max_price"); maxPriceStr != "" {
		maxPrice, err := strconv.ParseInt(maxPriceStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid max_price parameter: must be a number")
		}
		params.MaxPrice = &maxPrice
	}

	return params, nil
}
