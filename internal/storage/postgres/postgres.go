package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"example.com/market/internal/domain"
	"example.com/market/internal/storage"
	"github.com/jackc/pgerrcode"
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
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return storage.ErrUserExists
		}
		return fmt.Errorf("storage.CreateUser: %w", err)
	}

	return nil
}

// FindByLogin finds a user by their login.
func (s *Storage) FindByLogin(ctx context.Context, login string) (*domain.User, error) {
	const q = `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`

	rows, err := s.pool.Query(ctx, q, login)
	if err != nil {
		return nil, fmt.Errorf("storage.FindByLogin: %w", err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("storage.FindByLogin: %w", err)
	}

	return &u, nil
}

// FindUserByID finds a user by their ID.
func (s *Storage) FindUserByID(ctx context.Context, id int64) (*domain.User, error) {
	q := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	rows, err := s.pool.Query(ctx, q, id)
	if err != nil {
		return nil, fmt.Errorf("storage.FindUserByID: %w", err)
	}

	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByNameLax[domain.User])
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, storage.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("storage.FindUserByID: %w", err)
	}

	return &u, nil
}

// CreateAd creates a new ad in the database.
func (s *Storage) CreateAd(ctx context.Context, ad *domain.Ad) (int64, error) {
	q := `INSERT INTO ads (user_id, author_login, title, text, image_url, price) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at`

	err := s.pool.QueryRow(ctx, q, ad.UserID, ad.AuthorLogin, ad.Title, ad.Text, ad.ImageURL, ad.Price).Scan(&ad.ID, &ad.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return 0, storage.ErrForeignKeyViolation
			case pgerrcode.UniqueViolation:
				return 0, storage.ErrAdExists // Assuming a unique constraint on title or something similar
			}
		}
		return 0, fmt.Errorf("storage.CreateAd: %w", err)
	}

	return ad.ID, nil
}


// ListAds returns a list of ads with pagination and filtering.
func (s *Storage) ListAds(ctx context.Context, params *domain.ListAdsParams) ([]domain.Ad, error) {
	if params == nil {
		return nil, fmt.Errorf("storage.ListAds: params cannot be nil")
	}

	// Build the WHERE clause for price filtering
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if params.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("price >= $%d", argIndex))
		args = append(args, *params.MinPrice)
		argIndex++
	}

	if params.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("price <= $%d", argIndex))
		args = append(args, *params.MaxPrice)
		argIndex++
	}

	// Build the base query
	q := "SELECT id, user_id, author_login, title, text, image_url, price, created_at FROM ads"

	// Add WHERE clause if there are conditions
	if len(whereConditions) > 0 {
		q += " WHERE " + fmt.Sprintf("(%s)", fmt.Sprintf("%s", whereConditions[0]))
		for i := 1; i < len(whereConditions); i++ {
			q += " AND " + whereConditions[i]
		}
	}

	// Add ORDER BY clause
	q += fmt.Sprintf(" ORDER BY %s %s", params.SortBy, params.Order)

	// Add LIMIT and OFFSET for pagination
	q += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, params.GetOffset())

	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("storage.ListAds: %w", err)
	}

	ads, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[domain.Ad])
	if err != nil {
		return nil, fmt.Errorf("storage.ListAds: %w", err)
	}

	return ads, nil
}
