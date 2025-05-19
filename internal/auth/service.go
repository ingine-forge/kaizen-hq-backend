package auth

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/config"
	user "kaizen-hq/internal/account"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUserAlreadyExists = errors.New("user with this Torn ID already exists")

type Service struct {
	userService *user.Service
	config      *config.Config
}

func NewService(userService *user.Service, cfg *config.Config) *Service {
	return &Service{userService: userService, config: cfg}
}

func (s *Service) Register(ctx context.Context, user *user.User) error {
	// Check if user already exists
	_, err := s.userService.CreateUser(ctx, user)
	return err
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (string, error) {
	user, err := s.userService.GetUserByEmail(ctx, req.Email)
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
