package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) error {
	query := `INSERT INTO users (torn_id, username, password, api_key, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query, user.TornID, user.Username, user.Password, user.APIKey, time.Now())

	return err
}

// GetUserByTornID finds a user by their TornID
func (r *Repository) GetUserByTornID(ctx context.Context, tornID int) (*User, error) {
	user := &User{}

	query := `SELECT torn_id, username, password, api_key, created_at FROM users WHERE torn_id = $1`
	err := r.db.QueryRow(ctx, query, tornID).Scan(
		&user.TornID,
		&user.Username,
		&user.Password,
		&user.APIKey,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByUsername finds a user by their username
func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user := &User{}

	query := `SELECT torn_id, username, password, api_key, created_at FROM users WHERE username = $1`

	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.TornID,
		&user.Username,
		&user.Password,
		&user.APIKey,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}
