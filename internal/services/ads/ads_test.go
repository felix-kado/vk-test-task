package ads

import (
	"context"
	"errors"
	"testing"

	"example.com/market/internal/domain"
	"github.com/stretchr/testify/assert"
)

// mockAdRepository is a mock implementation of AdRepository for testing.
type mockAdRepository struct {
	CreateAdFunc   func(ctx context.Context, ad *domain.Ad) (int64, error)
	FindAdByIDFunc func(ctx context.Context, id int64) (*domain.Ad, error)
	ListAdsFunc    func(ctx context.Context, sortBy, order string) ([]domain.Ad, error)
}

func (m *mockAdRepository) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	return m.CreateAdFunc(ctx, ad)
}

func (m *mockAdRepository) FindAdByID(ctx context.Context, id int64) (*domain.Ad, error) {
	return m.FindAdByIDFunc(ctx, id)
}

func (m *mockAdRepository) ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
	return m.ListAdsFunc(ctx, sortBy, order)
}

func TestService_CreateAd(t *testing.T) {
	tests := []struct {
		name          string
		ad            *domain.Ad
		mockRepo      *mockAdRepository
		expectedID    int64
		expectedErr   error
	}{
		{
			name: "Success",
			ad: &domain.Ad{Title: "New Ad", Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{
				CreateAdFunc: func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 1, nil
				},
			},
			expectedID:  1,
			expectedErr: nil,
		},
		{
			name: "Validation Error - Title too long",
			ad: &domain.Ad{Title: string(make([]byte, 121)), Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{},
			expectedID:  0,
			expectedErr: ErrValidation,
		},
		{
			name: "Validation Error - Empty text",
			ad: &domain.Ad{Title: "New Ad", Text: "", UserID: 1},
			mockRepo: &mockAdRepository{},
			expectedID:  0,
			expectedErr: ErrValidation,
		},
		{
			name: "Repository Error",
			ad: &domain.Ad{Title: "New Ad", Text: "Some text", UserID: 1},
			mockRepo: &mockAdRepository{
				CreateAdFunc: func(ctx context.Context, ad *domain.Ad) (int64, error) {
					return 0, errors.New("db error")
				},
			},
			expectedID:  0,
			expectedErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := New(tt.mockRepo)
			id, err := service.CreateAd(context.Background(), tt.ad)

			assert.Equal(t, tt.expectedID, id)
			if tt.expectedErr != nil {
				assert.Error(t, err)
				// Use errors.Is for wrapped errors
				if !errors.Is(err, tt.expectedErr) && err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error '%v', got '%v'", tt.expectedErr, err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
