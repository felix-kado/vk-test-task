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

	"github.com/felix-kado/vk-test-task/internal/domain"
	"github.com/felix-kado/vk-test-task/internal/middleware"
	"github.com/felix-kado/vk-test-task/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockAdsService is a mock implementation of AdsService for testing.
type mockAdsService struct {
	CreateAdFunc func(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAdsFunc  func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error)
}

func (m *mockAdsService) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	return m.CreateAdFunc(ctx, ad)
}

func (m *mockAdsService) ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
	if m.ListAdsFunc != nil {
		return m.ListAdsFunc(ctx, params)
	}
	return nil, nil
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
			name:   "successful ad creation",
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
	tests := []struct {
		name                 string
		queryParams          string
		userID               int64 // For setting user in context
		setupMock            func(*mockAdsService)
		expectedStatus       int
		expectedBody         string
		expectedBodyContains []string
	}{
		{
			name:        "Success - authorized user sees ownership",
			queryParams: "",
			userID:      1, // Authenticated user with ID 1
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return []domain.Ad{
						{ID: 101, UserID: 1, Title: "My Own Ad", AuthorLogin: "test_user_1"},
						{ID: 102, UserID: 2, Title: "Someone Else's Ad", AuthorLogin: "test_user_2"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: []string{
				`"id":101`, `"title":"My Own Ad"`, `"is_owner":true`, `"author_login":"test_user_1"`,
				`"id":102`, `"title":"Someone Else's Ad"`, `"is_owner":false`, `"author_login":"test_user_2"`,
			},
		},
		{
			name:        "Success - unauthorized user does not see ownership",
			queryParams: "",
			userID:      0, // No user in context
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return []domain.Ad{
						{ID: 101, UserID: 1, Title: "An Ad", AuthorLogin: "test_user_1"},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBodyContains: []string{
				`"id":101`, `"title":"An Ad"`, `"is_owner":false`, `"author_login":"test_user_1"`,
			},
		},
		{
			name:        "validation error from service",
			queryParams: "?sort_by=invalid_field",
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return nil, fmt.Errorf("%w: invalid sort_by value", services.ErrInvalidInput)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid input: invalid sort_by value"}`,
		},
		{
			name:        "Success with pagination and filtering",
			queryParams: "?page=2&limit=5&min_price=100&max_price=500",
			userID:      1,
			setupMock: func(m *mockAdsService) {
				m.ListAdsFunc = func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					assert.Equal(t, 2, params.Page)
					assert.Equal(t, 5, params.Limit)
					assert.Equal(t, int64(100), *params.MinPrice)
					assert.Equal(t, int64(500), *params.MaxPrice)
					return []domain.Ad{},
						nil
				}
			},
			expectedStatus:       http.StatusOK,
			expectedBodyContains: []string{"[]"},
		},
		{
			name:           "Invalid page parameter",
			queryParams:    "?page=abc",
			setupMock:      func(m *mockAdsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid page parameter: must be a number"}`,
		},
		{
			name:           "Invalid limit parameter",
			queryParams:    "?limit=abc",
			setupMock:      func(m *mockAdsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid limit parameter: must be a number"}`,
		},
		{
			name:           "Invalid min_price parameter",
			queryParams:    "?min_price=abc",
			setupMock:      func(m *mockAdsService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid min_price parameter: must be a number"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAdsService{}
			tt.setupMock(mockSvc)

			handler := NewAdsHandler(mockSvc, slog.Default())

			req := httptest.NewRequest(http.MethodGet, "/ads"+tt.queryParams, nil)
			// Add user ID to context if provided for the test case
			if tt.userID != 0 {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			handler.ListAds(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else if len(tt.expectedBodyContains) > 0 {
				bodyStr := rr.Body.String()
				for _, sub := range tt.expectedBodyContains {
					assert.Contains(t, bodyStr, sub)
				}
			}
		})
	}
}
