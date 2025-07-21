package dto

import (
	"time"

	"example.com/market/internal/domain"
)

// AdResponse is a DTO for the Ad model, including an ownership flag and author login.
type AdResponse struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	ImageURL    string    `json:"image_url"`
	Price       int64     `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	AuthorLogin string    `json:"author_login"`
	IsOwner     bool      `json:"is_owner"`
}

// ToAdResponse converts a domain.Ad to AdResponse DTO.
// currentUserID is used to determine ownership (0 for unauthenticated users).
// authorLogin should be provided from the storage layer.
func ToAdResponse(ad *domain.Ad, authorLogin string, currentUserID int64) *AdResponse {
	return &AdResponse{
		ID:          ad.ID,
		UserID:      ad.UserID,
		Title:       ad.Title,
		Text:        ad.Text,
		ImageURL:    ad.ImageURL,
		Price:       ad.Price,
		CreatedAt:   ad.CreatedAt,
		AuthorLogin: authorLogin,
		IsOwner:     currentUserID != 0 && currentUserID == ad.UserID,
	}
}

// ToAdResponseList converts a slice of domain.Ad to AdResponse DTOs.
// userLogins is a map of userID -> login for efficient lookup.
// currentUserID is used to determine ownership (0 for unauthenticated users).
func ToAdResponseList(ads []domain.Ad, userLogins map[int64]string, currentUserID int64) []*AdResponse {
	responses := make([]*AdResponse, len(ads))
	for i, ad := range ads {
		authorLogin := userLogins[ad.UserID]
		responses[i] = ToAdResponse(&ad, authorLogin, currentUserID)
	}
	return responses
}
