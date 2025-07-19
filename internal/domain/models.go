package domain

import "time"

type User struct {
	ID           int64
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}

type Ad struct {
	ID        int64
	UserID    int64
	Title     string
	Text      string
	ImageURL  string
	Price     int64 // cents
	CreatedAt time.Time
}
