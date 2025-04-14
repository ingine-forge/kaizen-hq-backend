package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	Pool *pgxpool.Pool
}

func (s *Store) CreateUser(tornID int, username, password, apiKey string) (*User, error) {
	// Hash the password before storing
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Insert user into database
	query := `INSERT INTO users (torn_id, username, password, api_key, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err = s.Pool.Exec(context.Background(), query, tornID, username, hashedPassword, apiKey, time.Now())

	if err != nil {
		return nil, err
	}

	// Return the created user (without the password)
	return &User{
		TornID:    tornID,
		Username:  username,
		APIKey:    apiKey,
		CreatedAt: time.Now(),
	}, nil
}

// GetUserByTornID finds a user by their TornID
func (s *Store) GetUserByTornID(tornID int) (*User, error) {
	user := &User{}

	query := `SELECT torn_id, username, password, api_key, created_at FROM users WHERE torn_id = $1`
	err := s.Pool.QueryRow(context.Background(), query, tornID).Scan(
		&user.TornID,
		&user.Username,
		&user.Password,
		&user.APIKey,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// GetUserByUsername finds a user by their username
func (s *Store) GetUserByUsername(username string) (*User, error) {
	user := &User{}

	query := `SELECT torn_id, username, password, api_key, created_at FROM users WHERE username = $1`

	err := s.Pool.QueryRow(context.Background(), query, username).Scan(
		&user.TornID,
		&user.Username,
		&user.Password,
		&user.APIKey,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
