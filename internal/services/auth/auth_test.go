package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"example.com/market/internal/domain"
	"example.com/market/internal/services"
	"example.com/market/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// mockUserRepository is a mock implementation of UserRepository for testing.
type mockUserRepository struct {
	CreateUserFunc   func(ctx context.Context, u *domain.User) error
	FindByLoginFunc  func(ctx context.Context, login string) (*domain.User, error)
	FindUserByIDFunc func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, u *domain.User) error {
	return m.CreateUserFunc(ctx, u)
}

func (m *mockUserRepository) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	return m.FindByLoginFunc(ctx, login)
}

func (m *mockUserRepository) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	return m.FindUserByIDFunc(ctx, id)
}

func TestService_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			CreateUserFunc: func(ctx context.Context, u *domain.User) error {
				u.ID = 1 // Simulate DB assigning an ID
				return nil
			},
		}

		service := New(mockRepo, "test-secret", time.Hour)

		token, user, err := service.Register(context.Background(), "newuser", "ValidPass123!")

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		require.NotNil(t, user)
		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "newuser", user.Login)
	})

	t.Run("user already exists", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			CreateUserFunc: func(ctx context.Context, u *domain.User) error {
				return storage.ErrUserExists
			},
		}

		service := New(mockRepo, "test-secret", time.Hour)

		_, _, err := service.Register(context.Background(), "existinguser", "ValidPass123!")

		assert.Error(t, err)
		assert.True(t, errors.Is(err, services.ErrUserExists))
	})
}

func TestService_Register_Validation(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  string
	}{
		{"invalid login too short", "a", "ValidPass123!", ErrInvalidLogin.Error()},
		{"invalid login starts with number", "1user", "ValidPass123!", ErrInvalidLogin.Error()},
		{"invalid login with special chars", "user!", "ValidPass123!", ErrInvalidLogin.Error()},
		{"invalid password too short", "newuser", "short", ErrInvalidPassword.Error()},
		{"invalid password no uppercase", "newuser", "validpass123!", ErrInvalidPassword.Error()},
		{"invalid password no lowercase", "newuser", "VALIDPASS123!", ErrInvalidPassword.Error()},
		{"invalid password no number", "newuser", "ValidPass!", ErrInvalidPassword.Error()},
		{"invalid password no special char", "newuser", "ValidPass123", ErrInvalidPassword.Error()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			service := New(mockRepo, "test-secret", time.Hour)

			_, _, err := service.Register(context.Background(), tt.login, tt.password)

			require.Error(t, err)
			assert.True(t, errors.Is(err, services.ErrInvalidInput))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("ValidPass123!"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		mockRepo    *mockUserRepository
		login       string
		password    string
		expectToken bool
		expectedErr error
	}{
		{
			name: "Success",
			mockRepo: &mockUserRepository{
				FindByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}, nil
				},
			},
			login:       "testuser",
			password:    "ValidPass123!",
			expectToken: true,
			expectedErr: nil,
		},
		{
			name: "User not found",
			mockRepo: &mockUserRepository{
				FindByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
					return nil, storage.ErrUserNotFound
				},
			},
			login:       "nonexistent",
			password:    "ValidPass123!",
			expectToken: false,
			expectedErr: services.ErrInvalidCredentials,
		},
		{
			name: "Invalid password",
			mockRepo: &mockUserRepository{
				FindByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}, nil
				},
			},
			login:       "testuser",
			password:    "wrongpassword",
			expectToken: false,
			expectedErr: services.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo, "test-secret", time.Hour)
			token, err := service.Login(context.Background(), tt.login, tt.password)

			if tt.expectToken {
				assert.NotEmpty(t, token)
				assert.NoError(t, err)
			} else {
				assert.Empty(t, token)
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr), fmt.Sprintf("expected error %v, got %v", tt.expectedErr, err))
			}
		})
	}
}

func TestService_Login_Validation(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		password string
		wantErr  string
	}{
		{"invalid login", "a", "ValidPass123!", ErrInvalidLogin.Error()},
		{"empty password", "testuser", "", "password is required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			service := New(mockRepo, "test-secret", time.Hour)

			_, err := service.Login(context.Background(), tt.login, tt.password)

			require.Error(t, err)
			assert.True(t, errors.Is(err, services.ErrInvalidInput))
			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}
