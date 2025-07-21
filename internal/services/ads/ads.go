package ads

import (
	"context"
	"errors"
	"fmt"

	"example.com/market/internal/domain"
	"example.com/market/internal/services"
	"example.com/market/internal/storage"
)



// AdRepository defines the interface for ad storage.
type AdRepository interface {
	CreateAd(ctx context.Context, ad *domain.Ad) (int64, error)
	FindAdByID(ctx context.Context, id int64) (*domain.Ad, error)
	ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
}

// UserRepository defines the interface for user-related operations needed by ads service.
type UserRepository interface {
	GetUserLogins(ctx context.Context, userIDs []int64) (map[int64]string, error)
}

// Service provides ad-related operations.
type Service struct {
	adRepo  AdRepository
	userRepo UserRepository
}

// New creates a new ad service.
func New(adRepo AdRepository, userRepo UserRepository) *Service {
	return &Service{
		adRepo:   adRepo,
		userRepo: userRepo,
	}
}

// CreateAd creates a new ad after validating it.
func (s *Service) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	if err := s.validateAd(ad); err != nil {
		return 0, fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}

	adID, err := s.adRepo.CreateAd(ctx, ad)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidUserReference) {
			return 0, services.ErrForbidden
		}
		return 0, fmt.Errorf("adRepo.CreateAd: %w", err)
	}
	return adID, nil
}

func (s *Service) validateAd(ad *domain.Ad) error {
	if ad.Title == "" {
		return errors.New("title is required")
	}
	if len(ad.Title) > 120 {
		return errors.New("title is too long")
	}
	if ad.Text == "" {
		return errors.New("text cannot be empty")
	}
	if ad.UserID == 0 {
		return errors.New("user ID is required")
	}
	if ad.Price < 0 {
		return errors.New("price must be non-negative")
	}
	return nil
}

// ListAds returns a sorted list of ads.
func (s *Service) ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
	if err := validateListParams(sortBy, order); err != nil {
		return nil, fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}

	ads, err := s.adRepo.ListAds(ctx, sortBy, order)
	if err != nil {
		return nil, fmt.Errorf("ads.ListAds: %w", err)
	}
	return ads, nil
}

func validateListParams(sortBy, order string) error {
	if sortBy != "price" && sortBy != "created_at" {
		return errors.New("invalid sort_by parameter")
	}
	if order != "asc" && order != "desc" {
		return errors.New("invalid order parameter")
	}
	return nil
}

// GetAd returns an ad by its ID.
func (s *Service) GetAd(ctx context.Context, adID int64) (*domain.Ad, error) {
	ad, err := s.adRepo.FindAdByID(ctx, adID)
	if err != nil {
		if errors.Is(err, storage.ErrAdNotFound) {
			return nil, services.ErrAdNotFound
		}
		return nil, fmt.Errorf("adRepo.FindAdByID: %w", err)
	}
	return ad, nil
}

// GetUserLogins retrieves a map of user logins by their IDs.
func (s *Service) GetUserLogins(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	logins, err := s.userRepo.GetUserLogins(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("userRepo.GetUserLogins: %w", err)
	}
	return logins, nil
}
