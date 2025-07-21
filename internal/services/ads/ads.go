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
	ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error)
}

// UserRepository defines the interface for user-related operations needed by ads service.
type UserRepository interface {
	FindUserByID(ctx context.Context, id int64) (*domain.User, error)
}

// Service provides ad-related operations.
type Service struct {
	adRepo   AdRepository
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

	// Fetch user to get the login for denormalization
	user, err := s.userRepo.FindUserByID(ctx, ad.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			// This case should be handled by the foreign key constraint in the DB,
			// but checking here provides a clearer error.
			return 0, fmt.Errorf("%w: user not found", services.ErrInvalidInput)
		}
		return 0, fmt.Errorf("userRepo.FindUserByID: %w", err)
	}
	ad.AuthorLogin = user.Login

	adID, err := s.adRepo.CreateAd(ctx, ad)
	if err != nil {
		if errors.Is(err, storage.ErrForeignKeyViolation) {
			// This error is now less likely to be triggered by a missing user, but we keep it for other FK constraints.
			return 0, fmt.Errorf("%w: invalid data reference", services.ErrInvalidInput)
		}
		if errors.Is(err, storage.ErrAdExists) {
			return 0, services.ErrConflict
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

// ListAds returns a sorted and filtered list of ads with pagination.
func (s *Service) ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
	if err := s.validateListParams(params); err != nil {
		return nil, fmt.Errorf("%w: %v", services.ErrInvalidInput, err)
	}

	// Set defaults for unspecified parameters
	params.SetDefaults()

	ads, err := s.adRepo.ListAds(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("ads.ListAds: %w", err)
	}
	return ads, nil
}

// validateListParams validates the parameters for listing ads.
func (s *Service) validateListParams(params *domain.ListAdsParams) error {
	if params == nil {
		return errors.New("params cannot be nil")
	}

	// Validate sort_by if specified
	if params.SortBy != "" && params.SortBy != "price" && params.SortBy != "created_at" {
		return errors.New("invalid sort_by parameter: must be 'price' or 'created_at'")
	}

	// Validate order if specified
	if params.Order != "" && params.Order != "asc" && params.Order != "desc" {
		return errors.New("invalid order parameter: must be 'asc' or 'desc'")
	}

	// Validate pagination parameters
	if params.Page < 0 {
		return errors.New("page must be positive")
	}
	if params.Limit < 0 {
		return errors.New("limit must be positive")
	}
	if params.Limit > 100 {
		return errors.New("limit cannot exceed 100")
	}

	// Validate price filters
	if params.MinPrice != nil && *params.MinPrice < 0 {
		return errors.New("min_price must be non-negative")
	}
	if params.MaxPrice != nil && *params.MaxPrice < 0 {
		return errors.New("max_price must be non-negative")
	}
	if params.MinPrice != nil && params.MaxPrice != nil && *params.MinPrice > *params.MaxPrice {
		return errors.New("min_price cannot be greater than max_price")
	}

	return nil
}
