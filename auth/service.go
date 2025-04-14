package auth

import "errors"

type AuthService struct {
	store *Store
}

type LoginResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

func NewAuthService(store *Store) *AuthService {
	return &AuthService{store: store}
}

func (s *AuthService) Register(tornID int, username, password, apiKey string) (*User, error) {
	// Check if user already exists
	_, err := s.store.GetUserByTornID(tornID)
	if err == nil {
		return nil, errors.New("user with this Torn ID already exists")
	}

	// Create the user
	return s.store.CreateUser(tornID, username, password, apiKey)
}

func (s *AuthService) Login(username, password string) (*LoginResponse, error) {
	// Get the user
	user, err := s.store.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check password
	if !CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid username or password")
	}

	// Generate JWT token
	token, err := GenerateToken(user)
	if err != nil {
		return nil, err
	}

	// Don't return the password
	user.Password = ""

	// Return user and token
	return &LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *AuthService) GetUserByTornID(tornID int) (*User, error) {
	user, err := s.store.GetUserByTornID(tornID)
	if err != nil {
		return nil, err
	}

	// Don't return the password
	user.Password = ""
	return user, nil
}
