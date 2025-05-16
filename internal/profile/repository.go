package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) StoreProfile(ctx context.Context, profile Profile) error {
	fmt.Println(profile)
	query := `INSERT INTO profiles (torn_id, name, rank, property, donator, profile_image, signup, awards, level, friends, enemies, discord) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING torn_id`

	err := r.db.QueryRow(ctx, query, profile.TornID, profile.Name, profile.Rank, profile.Property, profile.Donator, profile.ProfileImage, profile.Signup, profile.Awards, profile.Level, profile.Friends, profile.Enemies, profile.Discord).Scan(&profile.TornID)

	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

func (r *Repository) GetProfileByTornID(ctx context.Context, id int) (*Profile, error) {
	profile := &Profile{}

	query := `SELECT * FROM profiles WHERE torn_id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&profile)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("profile not found")
		}
	}

	return profile, nil
}

func (r *Repository) GetProfileByDiscordID(ctx context.Context, id string) (*Profile, error) {
	profile := &Profile{}

	query := `SELECT name, rank, property, donator, torn_id, profile_image, signup, awards, level, friends, enemies, discord FROM profiles WHERE discord = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&profile.Name,
		&profile.Rank,
		&profile.Property,
		&profile.Donator,
		&profile.TornID,
		&profile.ProfileImage,
		&profile.Signup,
		&profile.Awards,
		&profile.Level,
		&profile.Friends,
		&profile.Enemies,
		&profile.Discord,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("profile not found")
		}
	}

	return profile, nil
}
