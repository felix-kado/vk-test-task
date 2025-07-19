package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/market/internal/domain"
	"example.com/market/internal/storage"
	"github.com/stretchr/testify/assert"
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
	mockRepo := &mockUserRepository{
		CreateUserFunc: func(ctx context.Context, u *domain.User) error {
			u.ID = 1 // Simulate DB assigning an ID
			return nil
		},
	}

	service := New(mockRepo, "test-secret", time.Hour)

	token, user, err := service.Register(context.Background(), "newuser", "password123")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "newuser", user.Login)
}

func TestService_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name          string
		mockRepo      *mockUserRepository
		login         string
		password      string
		expectToken   bool
		expectedErr   error
	}{
		{
			name: "Success",
			mockRepo: &mockUserRepository{
				FindByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser", PasswordHash: string(hashedPassword)}, nil
				},
			},
			login:       "testuser",
			password:    "password123",
			expectToken: true,
			expectedErr: nil,
		},
		{
			name: "User not found",
			mockRepo: &mockUserRepository{
				FindByLoginFunc: func(ctx context.Context, login string) (*domain.User, error) {
					return nil, storage.ErrNotFound
				},
			},
			login:       "nonexistent",
			password:    "password123",
			expectToken: false,
			expectedErr: ErrInvalidCredentials,
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
			expectedErr: ErrInvalidCredentials,
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
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}
