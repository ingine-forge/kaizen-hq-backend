package permission

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

func (s *Service) Create(ctx context.Context, permission *Permission) (*Permission, error) {
	// Check if the permission already exists
	_, err := s.repo.GetPermissionByName(ctx, permission.Name)
	if err == nil {
		return nil, errors.New("permission already exists")
	}
	return s.repo.CreatePermission(ctx, permission)
}
