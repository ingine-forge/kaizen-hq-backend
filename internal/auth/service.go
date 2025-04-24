package auth

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/config"
	"kaizen-hq/internal/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrUserAlreadyExists = errors.New("user with this Torn ID already exists")

type Service struct {
	userRepo *user.Repository
	config   *config.Config
}

func NewService(userRepo *user.Repository, cfg *config.Config) *Service {
	return &Service{userRepo: userRepo, config: cfg}
}

func (s *Service) Register(ctx context.Context, user *user.User) error {
	// Check if user already exists
	_, err := s.userRepo.GetUserByTornID(ctx, user.TornID)
	if err == nil {
		return ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	return s.userRepo.CreateUser(ctx, user)
}

func (s *Service) Login(ctx context.Context, req *LoginRequest) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, req.Username)
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

func (s *Service) GetCurrentUser(ctx context.Context, tornID int) {
	requester := ctx.Value("user")
	fmt.Println(requester)
}
