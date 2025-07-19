package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"example.com/market/internal/domain"
	"example.com/market/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository defines the interface for user storage.
// This allows us to mock the storage layer in tests.
type UserRepository interface {
	CreateUser(ctx context.Context, u *domain.User) error
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	FindUserByID(ctx context.Context, id int64) (*domain.User, error)
}

// Service provides user authentication operations.
type Service struct {
	userRepo UserRepository
	secret   []byte
	tokenTTL time.Duration
}

// New creates a new auth service.
func New(userRepo UserRepository, secret string, tokenTTL time.Duration) *Service {
	return &Service{
		userRepo: userRepo,
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
	}
}

// Register creates a new user and returns a JWT token.
func (s *Service) Register(ctx context.Context, login, password string) (string, *domain.User, error) {
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &domain.User{
		Login:        login,
		PasswordHash: string(passHash),
	}

	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		return "", nil, err
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, u, nil
}

// Login authenticates a user and returns a JWT token.
func (s *Service) Login(ctx context.Context, login, password string) (string, error) {
	u, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}

// ParseToken parses a JWT token and returns the user associated with it.
func (s *Service) ParseToken(ctx context.Context, tokenStr string) (*domain.User, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, err := claims.GetSubject()
		if err != nil {
			return nil, fmt.Errorf("invalid subject in token: %w", err)
		}

		var userID int64
		if _, err := fmt.Sscanf(sub, "%d", &userID); err != nil {
			return nil, fmt.Errorf("failed to parse user ID from token subject: %w", err)
		}

		u, err := s.userRepo.FindUserByID(ctx, userID)
		if err != nil {
			return nil, ErrUserNotFound
		}
		return u, nil
	}

	return nil, errors.New("invalid token")
}

func (s *Service) generateToken(u *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", u.ID),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}
