package auth

import (
	"context"
	"errors"
	"kaizen-hq/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	repo   *Repository
	config *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, config: cfg}
}

func (s *Service) Register(ctx context.Context, user *User) error {
	// Hash password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (string, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Compare password
	if !CheckPasswordHash(req.Password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	claims := &Claims{
		TornID:   user.TornID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}
