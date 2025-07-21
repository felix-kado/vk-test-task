package auth

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"example.com/market/internal/domain"
	"example.com/market/internal/services"
	"example.com/market/internal/storage"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service provides user authentication operations.
type Service struct {
	userRepo storage.UserRepository
	secret   []byte
	tokenTTL time.Duration
}

// New creates a new auth service.
func New(userRepo storage.UserRepository, secret string, tokenTTL time.Duration) *Service {
	return &Service{
		userRepo: userRepo,
		secret:   []byte(secret),
		tokenTTL: tokenTTL,
	}
}

// Register creates a new user and returns a JWT token.
func (s *Service) Register(ctx context.Context, login, password string) (string, *domain.User, error) {
	if err := validateLogin(login); err != nil {
		return "", nil, fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}
	if err := validatePassword(password); err != nil {
		return "", nil, fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u := &domain.User{
		Login:        login,
		PasswordHash: string(passHash),
	}

	if err := s.userRepo.CreateUser(ctx, u); err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return "", nil, services.ErrUserExists
		}
		return "", nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(u)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, u, nil
}

// Login authenticates a user and returns a JWT token.
func (s *Service) Login(ctx context.Context, login, password string) (string, error) {
	if err := validateLogin(login); err != nil {
		return "", fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}
	if password == "" {
		return "", fmt.Errorf("%w: password is required", services.ErrInvalidInput)
	}

	u, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return "", services.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return "", services.ErrInvalidCredentials
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
			if errors.Is(err, storage.ErrUserNotFound) {
				return nil, services.ErrUserNotFound
			}
			return nil, fmt.Errorf("failed to find user by id: %w", err)
		}
		return u, nil
	}

	return nil, errors.New("invalid token")
}

var (
	ErrInvalidLogin    = errors.New("login must be 3-50 characters long, start with a letter, and contain only letters, numbers, and underscores")
	ErrInvalidPassword = errors.New("password must be 8-72 characters long and contain at least one uppercase letter, one lowercase letter, one number, and one special character")
)

// validateLogin checks if the login meets the requirements
func validateLogin(login string) error {
	// Check length
	if len(login) < 3 || len(login) > 50 {
		return ErrInvalidLogin
	}

	// Check first character is a letter
	if !unicode.IsLetter(rune(login[0])) {
		return ErrInvalidLogin
	}

	// Check allowed characters (letters, numbers, underscore)
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, login)
	if !matched {
		return ErrInvalidLogin
	}

	return nil
}

// validatePassword checks if the password meets the requirements
func validatePassword(password string) error {
	// Check length
	if len(password) < 8 || len(password) > 72 {
		return ErrInvalidPassword
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsNumber(c):
			hasNumber = true
		case isSpecialCharacter(c):
			hasSpecial = true
		}

		// If all requirements are met, we can break early
		if hasUpper && hasLower && hasNumber && hasSpecial {
			break
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return ErrInvalidPassword
	}

	return nil
}

func isSpecialCharacter(r rune) bool {
	specialChars := "!@#$%^&*"
	for _, c := range specialChars {
		if r == c {
			return true
		}
	}
	return false
}

func (s *Service) generateToken(u *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": strconv.FormatInt(u.ID, 10),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(s.tokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}
