package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/market/internal/domain"
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
	mockService := &mockAuthService{
		RegisterFunc: func(ctx context.Context, login, password string) (string, *domain.User, error) {
			return "some-token", &domain.User{ID: 1, Login: login}, nil
		},
	}

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	handler := NewAuthHandler(mockService, logger)

	// Prepare request
	regReq := RegistrationRequest{Login: "test", Password: "password"}
	body, _ := json.Marshal(regReq)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]int64
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), resp["user_id"])
}

func TestAuthHandler_Login(t *testing.T) {
	mockService := &mockAuthService{
		LoginFunc: func(ctx context.Context, login, password string) (string, error) {
			return "a-valid-jwt-token", nil
		},
	}

	logger := slog.New(slog.NewTextHandler(bytes.NewBuffer(nil), nil))
	handler := NewAuthHandler(mockService, logger)

	// Prepare request
	loginReq := LoginRequest{Login: "test", Password: "password"}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Record response
	rr := httptest.NewRecorder()
	handler.Login(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "a-valid-jwt-token", resp["token"])
}
