package user

import (
	"kaizen-hq/config"
	"kaizen-hq/internal/client/torn"
)

type Service struct {
	repo       *Repository
	config     *config.Config
	tornClient torn.Client
}

func NewService(repo *Repository, cfg *config.Config, tornClient torn.Client) *Service {
	return &Service{repo: repo, config: cfg, tornClient: tornClient}
}
