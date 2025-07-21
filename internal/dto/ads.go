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
func ToAdResponse(ad *domain.Ad, currentUserID int64) *AdResponse {
	return &AdResponse{
		ID:          ad.ID,
		UserID:      ad.UserID,
		Title:       ad.Title,
		Text:        ad.Text,
		ImageURL:    ad.ImageURL,
		Price:       ad.Price,
		CreatedAt:   ad.CreatedAt,
		AuthorLogin: ad.AuthorLogin,
		IsOwner:     currentUserID != 0 && currentUserID == ad.UserID,
	}
}

// ToAdResponseList converts a slice of domain.Ad to AdResponse DTOs.
// currentUserID is used to determine ownership (0 for unauthenticated users).
func ToAdResponseList(ads []domain.Ad, currentUserID int64) []*AdResponse {
	responses := make([]*AdResponse, len(ads))
	for i := range ads {
		responses[i] = ToAdResponse(&ads[i], currentUserID)
	}
	return responses
}
