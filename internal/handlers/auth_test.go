package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/market/internal/domain"
	"example.com/market/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			expectedBody:   `{"user_id":1}`,
		},
		{
			name: "invalid login format - too short",
			request: map[string]string{
				"login":    "ab",
				"password": "ValidPass123!",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores`,
		},
		{
			name: "invalid login format - invalid characters",
			request: map[string]string{
				"login":    "test@user",
				"password": "ValidPass123!",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores`,
		},
		{
			name: "invalid password - too short",
			request: map[string]string{
				"login":    "testuser",
				"password": "short",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character`,
		},
		{
			name: "invalid password - missing requirements",
			request: map[string]string{
				"login":    "testuser",
				"password": "alllowercase",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character`,
		},
		{
			name: "user already exists",
			request: map[string]string{
				"login":    "existinguser",
				"password": "ValidPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.RegisterFunc = func(ctx context.Context, login, password string) (string, *domain.User, error) {
					return "", nil, storage.ErrExists
				}
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `user already exists`,
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
			assert.Contains(t, rr.Body.String(), tt.expectedBody)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
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
			name: "invalid login format",
			request: map[string]string{
				"login":    "ab",
				"password": "ValidPass123!",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores`,
		},
		{
			name: "missing password",
			request: map[string]string{
				"login":    "testuser",
				"password": "",
			},
			setupMock:      func(m *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `password is required`,
		},
		{
			name: "invalid credentials",
			request: map[string]string{
				"login":    "testuser",
				"password": "WrongPass123!",
			},
			setupMock: func(m *mockAuthService) {
				m.LoginFunc = func(ctx context.Context, login, password string) (string, error) {
					return "", storage.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `invalid credentials`,
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
			if tt.expectedStatus == http.StatusOK {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid login",
			login:   "user123",
			wantErr: false,
		},
		{
			name:    "login too short",
			login:   "ab",
			wantErr: true,
			errMsg:  "login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores",
		},
		{
			name:    "login with special chars",
			login:   "user@test",
			wantErr: true,
			errMsg:  "login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLogin(tt.login)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

	passwordTests := []struct {
		name    string
		pass    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid password",
			pass:    "ValidPass123!",
			wantErr: false,
		},
		{
			name:    "password too short",
			pass:    "Short1!",
			wantErr: true,
			errMsg:  "password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character",
		},
		{
			name:    "missing uppercase",
			pass:    "nopass123!",
			wantErr: true,
			errMsg:  "password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character",
		},
		{
			name:    "missing special char",
			pass:    "NoSpecial123",
			wantErr: true,
			errMsg:  "password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character",
		},
	}

	for _, tt := range passwordTests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.pass)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
