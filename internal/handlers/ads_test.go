package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/market/internal/domain"
	"example.com/market/internal/middleware"
	"example.com/market/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockAdsService is a mock implementation of AdsService for testing.
type mockAdsService struct {
	CreateAdFunc func(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAdsFunc  func(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
}

func (m *mockAdsService) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	return m.CreateAdFunc(ctx, ad)
}

func (m *mockAdsService) ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
	return m.ListAdsFunc(ctx, sortBy, order)
}

func TestAdsHandler_CreateAd(t *testing.T) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name           string
		requestBody    any
		userID         int64
		setupMock      func(*mockAdsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful ad creation",
			userID: 1,
			requestBody: AdRequest{
				Title: "Test Ad",
				Text:  "This is a test ad.",
			},
			setupMock: func(m *mockAdsService) {
				m.CreateAdFunc = func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 123, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":123}`,
		},
		{
			name:           "invalid request body",
			userID:         1,
			requestBody:    "not json",
			setupMock:      func(m *mockAdsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name:   "validation error from service",
			userID: 1,
			requestBody: AdRequest{
				Title: "", // Invalid title
				Text:  "Valid text.",
			},
			setupMock: func(m *mockAdsService) {
				m.CreateAdFunc = func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 0, fmt.Errorf("%w: title cannot be empty", services.ErrInvalidInput)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid input: title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAdsService{}
			tt.setupMock(mockSvc)

			handler := NewAdsHandler(mockSvc, slog.Default())

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/ads", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler.CreateAd(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code >= 400 {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				var errResp errorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errResp)
				assert.NoError(t, err, "failed to unmarshal error response")
				assert.Equal(t, tt.expectedBody, errResp.Error)
			} else {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestAdsHandler_ListAds(t *testing.T) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*mockAdsService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful ad listing",
			queryParams: "?sort_by=price&order=asc",
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
					return []domain.Ad{{ID: 1, Title: "Ad 1"}}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"user_id":0,"title":"Ad 1","text":"","image_url":"","price":0,"created_at":"0001-01-01T00:00:00Z"}]`,
		},
		{
			name:        "validation error from service",
			queryParams: "?sort_by=invalid_field",
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
					return nil, fmt.Errorf("%w: invalid sort_by value", services.ErrInvalidInput)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid input: invalid sort_by value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAdsService{}
			tt.setupMock(mockSvc)

			handler := NewAdsHandler(mockSvc, slog.Default())

			req := httptest.NewRequest(http.MethodGet, "/ads"+tt.queryParams, nil)

			rr := httptest.NewRecorder()
			handler.ListAds(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code >= 400 {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				var errResp errorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errResp)
				assert.NoError(t, err, "failed to unmarshal error response")
				assert.Equal(t, tt.expectedBody, errResp.Error)
			} else {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			}
		})
	}
}
