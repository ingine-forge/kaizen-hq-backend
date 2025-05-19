package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAccount(ctx context.Context, account *Account) (int, error) {
	query := `INSERT INTO accounts (torn_id, email, password_hash, api_key, discord_id, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRow(ctx, query, account.TornID, account.Email, account.Password, account.APIKey, account.DiscordID, time.Now()).Scan(&account.ID)

	if err != nil {
		fmt.Println(err)
		return 0, fmt.Errorf("failed to create account: %w", err)
	}

	return account.ID, err
}

// GetUserByTornID finds a user by their TornID
func (r *Repository) GetAccountByTornID(ctx context.Context, tornID int) (*Account, error) {
	user := &Account{}

	query := `SELECT id, torn_id, email, api_key, created_at FROM users WHERE torn_id = $1`
	err := r.db.QueryRow(ctx, query, tornID).Scan(
		&user.ID,
		&user.TornID,
		&user.Email,
		&user.APIKey,
		&user.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserByEmail finds a user by their username
func (r *Repository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	user := &Account{}

	query := `SELECT torn_id, email, password, api_key, created_at FROM users WHERE email = $1`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.TornID,
		&user.Email,
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

// Count gives the number of users recorded in the database
func (r *Repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM accounts`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// AssignRole assigns role to a user
func (r *Repository) AssignRole(ctx context.Context, userID, roleID int) error {
	query := `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	_, err := r.db.Exec(ctx, query, userID, roleID)

	return err
}
