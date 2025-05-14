package role

import (
	"context"
	"errors"
	"kaizen-hq/config"
	"kaizen-hq/internal/permission"
)

type Service struct {
	repo   *Repository
	config *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, config: cfg}
}

func (s *Service) Create(ctx context.Context, role *Role) (*Role, error) {
	// Check if the role already exists
	_, err := s.repo.GetRoleByName(ctx, role.Name)
	if err == nil {
		return nil, errors.New("role already exists")
	}

	return s.repo.CreateRole(ctx, role)
}

func (s *Service) AssignPermission(ctx context.Context, role *Role, permission *permission.Permission) error {
	return s.repo.AssignPermission(ctx, role.ID, permission.ID)
}
