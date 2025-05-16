package user

import (
	"context"
	"errors"
	"kaizen-hq/config"
)

type Service struct {
	repo   *Repository
	config *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, config: cfg}
}

func (s *Service) GetUserByTornID(
	ctx context.Context,
	tornID int,
	currentUserTornID int,
) (*User, error) {
	user, err := s.repo.GetUserByTornID(ctx, tornID)
	if err != nil {
		return nil, err
	}

	if tornID != currentUserTornID {
		user.APIKey = ""
	}

	return user, nil
}

func (s *Service) GetUserByEmail(
	ctx context.Context,
	username string,
) (*User, error) {
	user, err := s.repo.GetUserByEmail(ctx, username)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *Service) CreateUser(ctx context.Context, user *User) (int, error) {
	// Check if user already exists
	_, err := s.repo.GetUserByTornID(ctx, user.TornID)
	if err == nil {
		return 0, errors.New("user with this Torn ID already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return 0, err
	}
	user.Password = hashedPassword

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) AssignRole(ctx context.Context, userID, roleID int) error {
	return s.repo.AssignRole(ctx, userID, roleID)
}
