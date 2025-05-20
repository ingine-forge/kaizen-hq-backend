package user

import (
	"context"
	"errors"
	"kaizen-hq/config"
	"kaizen-hq/internal/client"
	"strconv"
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

func (s *Service) CreateUser(ctx context.Context, playerID int, apiKey string) error {
	// Check if user already exists
	_, err := s.repo.GetUserByPlayerID(ctx, playerID)
	if err == nil {
		return errors.New("user with this Torn ID already exists")
	}

	// Fetch the user using torn client
	tornUser, err := s.tornClient.FetchTornUser(ctx, apiKey, strconv.Itoa(playerID))
	if err != nil {
		return errors.New("trouble fetching torn user")
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
		PlayerID:     playerID,
		Name:         tornUser.Name,
		PropertyID:   tornUser.PropertyID,
		Revivable:    tornUser.Revivable,
		ProfileImage: tornUser.ProfileImage,
	}

	return s.repo.CreateUser(ctx, user)
}
