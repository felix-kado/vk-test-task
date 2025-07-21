package storage

import (
	"context"

	"example.com/market/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *domain.User) error
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	FindUserByID(ctx context.Context, id int64) (*domain.User, error)
	// GetUserLogins returns a map of userID -> login for the given user IDs
	GetUserLogins(ctx context.Context, userIDs []int64) (map[int64]string, error)
}

type AdRepository interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	FindAdByID(ctx context.Context, id int64) (*domain.Ad, error)
	ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
	// ListAds will be implemented later with filtering and pagination
}
