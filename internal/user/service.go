package user

import (
	"context"
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
