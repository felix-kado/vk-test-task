package ads

import (
	"context"
	"errors"
	"fmt"

	"example.com/market/internal/domain"
)

var (
	ErrValidation = errors.New("validation failed")
)

// AdRepository defines the interface for ad storage.
type AdRepository interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	FindAdByID(ctx context.Context, id int64) (*domain.Ad, error)
	ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
}

// Service provides ad-related operations.
type Service struct {
	adRepo AdRepository
}

// New creates a new ad service.
func New(adRepo AdRepository) *Service {
	return &Service{adRepo: adRepo}
}

// CreateAd creates a new ad after validating it.
func (s *Service) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	if err := s.validateAd(ad); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	return s.adRepo.CreateAd(ctx, ad)
}

func (s *Service) validateAd(ad *domain.Ad) error {
	if len(ad.Title) > 120 {
		return errors.New("title is too long")
	}
	if ad.Text == "" {
		return errors.New("text cannot be empty")
	}
	if ad.UserID == 0 {
		return errors.New("user ID is required")
	}
	return nil
}

// ListAds returns a sorted list of ads.
func (s *Service) ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
	ads, err := s.adRepo.ListAds(ctx, sortBy, order)
	if err != nil {
		return nil, fmt.Errorf("ads.ListAds: %w", err)
	}
	return ads, nil
}
