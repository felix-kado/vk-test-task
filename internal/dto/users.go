package dto

import (
	"time"

	"github.com/felix-kado/vk-test-task/internal/domain"
)

// UserResponse is a DTO for the User model, used in registration responses.
// It excludes sensitive information like password hash.
type UserResponse struct {
	ID        int64     `json:"id"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"created_at"`
}

// ToUserResponse converts a domain.User to UserResponse DTO.
// It excludes sensitive information like password hash.
func ToUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Login:     user.Login,
		CreatedAt: user.CreatedAt,
	}
}
