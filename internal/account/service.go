package account

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

func (s *Service) GetAccountByTornID(
	ctx context.Context,
	tornID int,
	currentAccountTornID int,
) (*Account, error) {
	account, err := s.repo.GetAccountByTornID(ctx, tornID)
	if err != nil {
		return nil, err
	}

	if tornID != currentAccountTornID {
		account.APIKey = ""
	}

	return account, nil
}

func (s *Service) GetAccountByEmail(
	ctx context.Context,
	email string,
) (*Account, error) {
	user, err := s.repo.GetAccountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *Service) CreateAccount(ctx context.Context, account *Account) (int, error) {
	// Check if user already exists
	_, err := s.repo.GetAccountByTornID(ctx, account.TornID)
	if err == nil {
		return 0, errors.New("account with this Torn ID already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(account.Password)
	if err != nil {
		return 0, err
	}
	account.Password = hashedPassword

	return s.repo.CreateAccount(ctx, account)
}

func (s *Service) AssignRole(ctx context.Context, accountID, roleID int) error {
	return s.repo.AssignRole(ctx, accountID, roleID)
}
