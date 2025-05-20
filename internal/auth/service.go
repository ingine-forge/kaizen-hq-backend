package auth

import (
	"context"
	"errors"
	"fmt"
	"kaizen-hq/config"
	"kaizen-hq/internal/account"
	"kaizen-hq/internal/client"
	"kaizen-hq/internal/user"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	ErrEmailAlreadyRegistered = "user with this email is already registered"
	ErrInvalidAPIKey          = "the api key provided could not be verified"
	ErrInvalidAPIKeyAccess    = "the api key must be of either limited access or full access"
	ErrUserNotFound           = "trouble finding the user in torn"
	ErrUserAlreadyRegistered  = "the user with this api key is already registered"
	ErrAccountCreationFailed  = "failed to create new account"
)

type Service struct {
	accountService *account.Service
	userService    *user.Service
	config         *config.Config
	tornClient     client.Client
}

func NewService(accountService *account.Service, userService *user.Service, cfg *config.Config, tornClient client.Client) *Service {
	return &Service{accountService: accountService, userService: userService, config: cfg, tornClient: tornClient}
}

// Helper method to verify API key
func (s *Service) verifyAPIKey(ctx context.Context, apiKey string) (int, error) {
	accessLevel, err := s.tornClient.FetchKeyDetails(ctx, apiKey)
	if err != nil {
		return 0, fmt.Errorf("API key verification failed: %w", err)
	}
	return accessLevel, nil
}

// Helper method to fetch Torn user
func (s *Service) fetchTornUser(ctx context.Context, apiKey string) (*client.User, error) {
	user, err := s.tornClient.FetchTornUser(ctx, apiKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Torn user details: %w", err)
	}
	return user, nil
}

func (s *Service) Register(ctx context.Context, account *account.Account) error {
	// Check if the user with the email already exists
	if _, err := s.accountService.GetAccountByEmail(ctx, account.Email); err == nil {
		return errors.New(ErrEmailAlreadyRegistered)
	}

	// Fetch the torn user
	tornUser, err := s.fetchTornUser(ctx, account.APIKey)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrUserNotFound, err)
	}

	account.TornID = tornUser.PlayerID

	// Verify if API key is valid
	accessLevel, err := s.verifyAPIKey(ctx, account.APIKey)
	if err != nil {
		return fmt.Errorf("%s: %w", ErrInvalidAPIKey, err)
	}

	if accessLevel < 3 {
		return errors.New(ErrInvalidAPIKeyAccess)
	}

	// Check if account is already registered
	if _, err := s.accountService.GetAccountByTornID(ctx, tornUser.PlayerID, tornUser.PlayerID); err == nil {
		return fmt.Errorf(ErrUserAlreadyRegistered+": %d", tornUser.PlayerID)
	}

	// Pass Torn user directly to avoid second fetch
	if err := s.userService.CreateUserIfNotExists(ctx, tornUser); err != nil {
		return err
	}

	// Create account
	if _, err := s.accountService.CreateAccount(ctx, account); err != nil {
		return fmt.Errorf(ErrAccountCreationFailed+": %w", err)
	}

	return nil
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
