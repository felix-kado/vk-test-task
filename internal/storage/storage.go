package storage

import (
	"context"

	"example.com/market/internal/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, u *domain.User) error
	FindByLogin(ctx context.Context, login string) (*domain.User, error)
	FindUserByID(ctx context.Context, id int64) (*domain.User, error)
}

type AdRepository interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error)
}
