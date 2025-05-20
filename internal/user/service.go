package user

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/config"
	"kaizen-hq/internal/client"
)

type Service struct {
	repo       *Repository
	config     *config.Config
	tornClient client.Client
}

func NewService(repo *Repository, cfg *config.Config, tornClient client.Client) *Service {
	return &Service{repo: repo, config: cfg, tornClient: tornClient}
}

func (s *Service) GetUserByPlayerID(
	ctx context.Context,
	playerID int,
) (*User, error) {
	user, err := s.repo.GetUserByPlayerID(ctx, playerID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) CreateUserIfNotExists(ctx context.Context, tornUser *client.User) error {
	// Check if user already exists
	_, err := s.repo.GetUserByPlayerID(ctx, tornUser.PlayerID)
	if err == nil {
		return nil // No-op; user already exists
	}

	if !errors.Is(err, ErrProfileNotFound) {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}

	user := User{
		Rank:         tornUser.Rank,
		Level:        tornUser.Level,
		Honor:        tornUser.Honor,
		Gender:       tornUser.Gender,
		Property:     tornUser.Property,
		Signup:       tornUser.Signup,
		Awards:       tornUser.Awards,
		Friends:      tornUser.Friends,
		Enemies:      tornUser.Enemies,
		ForumPosts:   tornUser.ForumPosts,
		Karma:        tornUser.Karma,
		Age:          tornUser.Age,
		Role:         tornUser.Role,
		Donator:      tornUser.Donator,
		PlayerID:     tornUser.PlayerID,
		Name:         tornUser.Name,
		PropertyID:   tornUser.PropertyID,
		Revivable:    tornUser.Revivable,
		ProfileImage: tornUser.ProfileImage,
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) EnsureUserExists(ctx context.Context, tornID string, apiKey string) (*User, error) {
	tornUser, err := s.tornClient.FetchTornUser(ctx, apiKey, tornID)
	if err != nil {
		return nil, err
	}
	err = s.CreateUserIfNotExists(ctx, tornUser)
	if err != nil {
		return nil, err
	}
	return &User{PlayerID: tornUser.PlayerID, Name: tornUser.Name}, nil
}
