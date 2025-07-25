package ads

import (
	"context"
	"errors"
	"testing"

	"github.com/felix-kado/vk-test-task/internal/domain"
	"github.com/felix-kado/vk-test-task/internal/services"
	"github.com/stretchr/testify/assert"
)

// mockAdRepository is a mock implementation of AdRepository for testing.
type mockAdRepository struct {
	CreateAdFunc   func(ctx context.Context, ad *domain.Ad) (int64, error)
	ListAdsFunc    func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error)
}

// mockUserRepository is a mock implementation of UserRepository for testing.
type mockUserRepository struct {
	FindUserByIDFunc func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserRepository) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	if m.FindUserByIDFunc != nil {
		return m.FindUserByIDFunc(ctx, id)
	}
	return nil, errors.New("FindUserByIDFunc not implemented")
}

func (m *mockAdRepository) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	if m.CreateAdFunc != nil {
		return m.CreateAdFunc(ctx, ad)
	}
	return 0, nil
}

func (m *mockAdRepository) ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
	if m.ListAdsFunc != nil {
		return m.ListAdsFunc(ctx, params)
	}
	return nil, nil
}

func int64Ptr(i int64) *int64 {
	return &i
}

func TestService_CreateAd(t *testing.T) {
	tests := []struct {
		name         string
		ad           *domain.Ad
		mockRepo     *mockAdRepository
		mockUserRepo *mockUserRepository
		expectedID   int64
		expectedErr  error
	}{
		{
			name: "Success",
			ad:   &domain.Ad{Title: "New Ad", Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{
				CreateAdFunc: func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 1, nil
				},
			},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name:     "Validation Error - Title too long",
			ad:       &domain.Ad{Title: string(make([]byte, 121)), Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  0,
			expectedErr: services.ErrInvalidInput,
		},
		{
			name:     "Validation Error - Empty text",
			ad:       &domain.Ad{Title: "New Ad", Text: "", UserID: 1},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  0,
			expectedErr: services.ErrInvalidInput,
		},
		{
			name:     "Validation Error - Missing title",
			ad:       &domain.Ad{Title: "", Text: "Some text", UserID: 1, Price: 100},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  0,
			expectedErr: services.ErrInvalidInput,
		},
		{
			name:     "Validation Error - Negative price",
			ad:       &domain.Ad{Title: "New Ad", Text: "Some text", UserID: 1, Price: -100},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  0,
			expectedErr: services.ErrInvalidInput,
		},
		{
			name: "Repository Error",
			ad:   &domain.Ad{Title: "New Ad", Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{
				CreateAdFunc: func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 0, errors.New("db error")
				},
			},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedID:  0,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo, tt.mockUserRepo)
			id, err := service.CreateAd(context.Background(), tt.ad)

			assert.Equal(t, tt.expectedID, id)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				if errors.Is(err, services.ErrInvalidInput) {
					assert.ErrorIs(t, err, tt.expectedErr)
				} else {
					assert.ErrorContains(t, err, tt.expectedErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ListAds(t *testing.T) {
	tests := []struct {
		name         string
		params       *domain.ListAdsParams
		mockRepo     *mockAdRepository
		mockUserRepo *mockUserRepository
		expectedErr  error
	}{
		{
			name:   "Success",
			params: &domain.ListAdsParams{SortBy: "price", Order: "asc"},
			mockRepo: &mockAdRepository{
				ListAdsFunc: func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return []domain.Ad{}, nil
				},
			},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  nil,
		},
		{
			name:     "Invalid sort_by",
			params:   &domain.ListAdsParams{SortBy: "name", Order: "asc"},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedErr: services.ErrInvalidInput,
		},
		{
			name:     "Invalid order",
			params:   &domain.ListAdsParams{SortBy: "price", Order: "descending"},
			mockRepo: &mockAdRepository{},
			mockUserRepo: &mockUserRepository{
				FindUserByIDFunc: func(ctx context.Context, id int64) (*domain.User, error) {
					return &domain.User{ID: 1, Login: "testuser"}, nil
				},
			},
			expectedErr: services.ErrInvalidInput,
		},
		{
			name:   "Valid pagination",
			params: &domain.ListAdsParams{Page: 2, Limit: 5},
			mockRepo: &mockAdRepository{
				ListAdsFunc: func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return []domain.Ad{}, nil
				},
			},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  nil,
		},
		{
			name:   "Valid filtering",
			params: &domain.ListAdsParams{MinPrice: int64Ptr(100), MaxPrice: int64Ptr(500)},
			mockRepo: &mockAdRepository{
				ListAdsFunc: func(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
					return []domain.Ad{}, nil
				},
			},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  nil,
		},
		{
			name:         "Invalid page",
			params:       &domain.ListAdsParams{Page: -1},
			mockRepo:     &mockAdRepository{},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  services.ErrInvalidInput,
		},
		{
			name:         "Invalid limit",
			params:       &domain.ListAdsParams{Limit: -1},
			mockRepo:     &mockAdRepository{},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  services.ErrInvalidInput,
		},
		{
			name:         "Invalid price range",
			params:       &domain.ListAdsParams{MinPrice: int64Ptr(500), MaxPrice: int64Ptr(100)},
			mockRepo:     &mockAdRepository{},
			mockUserRepo: &mockUserRepository{},
			expectedErr:  services.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo, tt.mockUserRepo)
			ads, err := service.ListAds(context.Background(), tt.params)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr), "expected error '%v', got '%v'", tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ads)
			}
		})
	}
}
