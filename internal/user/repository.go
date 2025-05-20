package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProfileNotFound = errors.New("profile not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user User) error {
	query := `INSERT INTO users (rank, level, honor, gender, property, signup, awards, friends, enemies, forum_posts, karma, age, role, donator, player_id, name, property_id, revivable, profile_image) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19) RETURNING player_id`

	err := r.db.QueryRow(ctx, query, user.Rank, user.Level, user.Honor, user.Gender, user.Property, user.Signup, user.Awards, user.Friends, user.Enemies, user.ForumPosts, user.Karma, user.Age, user.Role, user.Donator, user.PlayerID, user.Name, user.PropertyID, user.Revivable, user.ProfileImage).Scan(&user.PlayerID)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *Repository) GetUserByPlayerID(ctx context.Context, id int) (*User, error) {
	user := &User{}

	query := `SELECT * FROM users WHERE player_id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&user)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrProfileNotFound
		}
	}

	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user User) error {
	query := `UPDATE users
           SET rank = $1, level = $2, honor = $3, gender = $4, property = $5,
               signup = $6, awards = $7, friends = $8, enemies = $9, forum_posts = $10,
               karma = $11, age = $12, role = $13, donator = $14, name = $15, property_id = $16, revivable = $17
           WHERE player_id = $18
           RETURNING player_id`

	params := []any{
		user.Rank, user.Level, user.Honor, user.Gender, user.Property, user.Signup,
		user.Awards, user.Friends, user.Enemies, user.ForumPosts, user.Karma, user.Age,
		user.Role, user.Donator, user.Name, user.PropertyID, user.Revivable, user.PlayerID,
	}

	err := r.db.QueryRow(ctx, query, params...).Scan(&user.PlayerID)
	if err != nil {
		return err
	}

	return nil
}
