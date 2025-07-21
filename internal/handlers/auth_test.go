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
	"github.com/felix-kado/vk-test-task/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockAuthService is a mock implementation of AuthService for testing.
type mockAuthService struct {
	RegisterFunc func(ctx context.Context, login, password string) (string, *domain.User, error)
	LoginFunc    func(ctx context.Context, login, password string) (string, error)
}

func (m *mockAuthService) Register(ctx context.Context, login, password string) (string, *domain.User, error) {
	return m.RegisterFunc(ctx, login, password)
}

func (m *mockAuthService) Login(ctx context.Context, login, password string) (string, error) {
	return m.LoginFunc(ctx, login, password)
}

func TestAuthHandler_Register(t *testing.T) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name           string
		request        map[string]string
		setupMock      func(*mockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful registration",
			request: map[string]string{
				"login":    "testuser",
				"password": "ValidPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.RegisterFunc = func(ctx context.Context, login, password string) (string, *domain.User, error) {
					return "token", &domain.User{ID: 1, Login: login}, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"id":1,"login":"testuser","created_at":"0001-01-01T00:00:00Z"}`,
		},
		{
			name: "validation error from service",
			request: map[string]string{
				"login":    "a",
				"password": "short",
			},
			setupMock: func(m *mockAuthService) {
				m.RegisterFunc = func(ctx context.Context, login, password string) (string, *domain.User, error) {
					return "", nil, fmt.Errorf("%w: invalid password", services.ErrInvalidInput)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid input: invalid password",
		},
		{
			name: "user already exists",
			request: map[string]string{
				"login":    "existinguser",
				"password": "ValidPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.RegisterFunc = func(ctx context.Context, login, password string) (string, *domain.User, error) {
					return "", nil, services.ErrUserExists
				}
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "user with this login already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{}
			tt.setupMock(mockSvc)

			handler := NewAuthHandler(mockSvc, slog.Default())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Register(rr, req)

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

func TestAuthHandler_Login(t *testing.T) {
	type errorResponse struct {
		Error string `json:"error"`
	}

	tests := []struct {
		name           string
		request        map[string]string
		setupMock      func(*mockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful login",
			request: map[string]string{
				"login":    "testuser",
				"password": "ValidPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.LoginFunc = func(ctx context.Context, login, password string) (string, error) {
					return "token", nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"token":"token"}`,
		},
		{
			name: "validation error from service",
			request: map[string]string{
				"login":    "a",
				"password": "ValidPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.LoginFunc = func(ctx context.Context, login, password string) (string, error) {
					return "", fmt.Errorf("%w: invalid login", services.ErrInvalidInput)
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid input: invalid login",
		},
		{
			name: "invalid credentials",
			request: map[string]string{
				"login":    "testuser",
				"password": "WrongPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.LoginFunc = func(ctx context.Context, login, password string) (string, error) {
					return "", services.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "invalid login or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockAuthService{}
			tt.setupMock(mockSvc)

			handler := NewAuthHandler(mockSvc, slog.Default())

			body, _ := json.Marshal(tt.request)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

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
