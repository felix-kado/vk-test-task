package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"example.com/market/internal/domain"
	"example.com/market/internal/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Storage implements the storage interfaces for PostgreSQL.
type Storage struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// New creates a new PostgreSQL storage instance and connects to the database.
func New(ctx context.Context, dsn string, log *slog.Logger) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Storage{pool: pool, log: log}, nil
}

// Close closes the database connection pool.
func (s *Storage) Close() {
	s.pool.Close()
}

// CreateUser creates a new user in the database.
func (s *Storage) CreateUser(ctx context.Context, u *domain.User) error {
	q := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, created_at`

	err := s.pool.QueryRow(ctx, q, u.Login, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return storage.ErrExists
		}
		return err
	}

	return nil
}

// FindByLogin finds a user by their login.
func (s *Storage) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	q := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	var u domain.User
	err := s.pool.QueryRow(ctx, q, login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("storage.FindByLogin: %w", err)
	}

	return &u, nil
}

// FindUserByID finds a user by their ID.
func (s *Storage) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	q := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	var u domain.User
	err := s.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("storage.FindUserByID: %w", err)
	}

	return &u, nil
}

// CreateAd creates a new ad in the database.
func (s *Storage) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	q := `INSERT INTO ads (user_id, title, text, image_url, price) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := s.pool.QueryRow(ctx, q, ad.UserID, ad.Title, ad.Text, ad.ImageURL, ad.Price).Scan(&ad.ID, &ad.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key violation on users table
			return 0, storage.ErrNotFound
		}
		return 0, fmt.Errorf("storage.CreateAd: %w", err)
	}

	return ad.ID, nil
}

// FindAdByID finds an ad by its ID.
func (s *Storage) FindAdByID(ctx context.Context, id int64) (*domain.Ad, error) {
	q := `SELECT id, user_id, title, text, image_url, price, created_at FROM ads WHERE id = $1`

	var ad domain.Ad
	err := s.pool.QueryRow(ctx, q, id).Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Text, &ad.ImageURL, &ad.Price, &ad.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, fmt.Errorf("storage.FindAdByID: %w", err)
	}

	return &ad, nil
}

// ListAds returns a list of ads, sorted by the given column and order.
func (s *Storage) ListAds(ctx context.Context, sortBy, order string) ([]domain.Ad, error) {
	// Whitelist sortable columns and order to prevent SQL injection
	if sortBy != "price" && sortBy != "created_at" {
		sortBy = "created_at" // default sort
	}
	if order != "asc" && order != "desc" {
		order = "desc" // default order
	}

	q := fmt.Sprintf(`SELECT id, user_id, title, text, image_url, price, created_at FROM ads ORDER BY %s %s`, sortBy, order)

	rows, err := s.pool.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("storage.ListAds: %w", err)
	}
	defer rows.Close()

	var ads []domain.Ad
	for rows.Next() {
		var ad domain.Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Title, &ad.Text, &ad.ImageURL, &ad.Price, &ad.CreatedAt); err != nil {
			return nil, fmt.Errorf("storage.ListAds: scan error: %w", err)
		}
		ads = append(ads, ad)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("storage.ListAds: rows error: %w", err)
	}

	return ads, nil
}
