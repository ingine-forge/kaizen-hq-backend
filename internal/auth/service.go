package auth

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/config"
	"kaizen-hq/internal/account"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUserAlreadyExists = errors.New("user with this Torn ID already exists")

type Service struct {
	accountService *account.Service
	config         *config.Config
}

func NewService(accountService *account.Service, cfg *config.Config) *Service {
	return &Service{accountService: accountService, config: cfg}
}

func (s *Service) Register(ctx context.Context, account *account.Account) error {
	// Check if user already exists
	_, err := s.accountService.CreateAccount(ctx, account)
	return err
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (string, error) {
	user, err := s.accountService.GetAccountByEmail(ctx, req.Email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Compare password
	if !CheckPasswordHash(req.Password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	claims := &Claims{
		TornID: user.TornID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

func (s *Service) GetCurrentUser(ctx context.Context, tornID int) {
	requester := ctx.Value("user")
	fmt.Println(requester)
}
