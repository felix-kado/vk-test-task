package domain

import "time"

type User struct {
	ID           int64     `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Ad struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Title       string    `json:"title"`
	Text        string    `json:"text"`
	ImageURL    string    `json:"image_url,omitempty"`
	Price       int64     `json:"price"`
	AuthorLogin string    `json:"author_login"`
	CreatedAt   time.Time `json:"created_at"`
}
