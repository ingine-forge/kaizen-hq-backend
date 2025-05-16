package profile

import (
	"context"
	"errors"
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

func (s *Service) StoreProfileForID(ctx context.Context, tornID int) error {
	tornProfile, err := s.tornClient.FetchUserProfile(ctx, tornID)

	if err != nil {
		return errors.New("could not fetch torn profile")
	}

	discordID, err := s.tornClient.FetchDiscordID(ctx, tornID)
	if err != nil {
		return errors.New("could not fetch discord ID")
	}

	profile := Profile{
		Name:         tornProfile.Name,
		Rank:         tornProfile.Rank,
		Property:     tornProfile.Property,
		Donator:      tornProfile.Donator,
		TornID:       tornProfile.PlayerID,
		ProfileImage: tornProfile.ProfileImage,
		Signup:       tornProfile.Signup,
		Awards:       tornProfile.Awards,
		Level:        tornProfile.Level,
		Friends:      tornProfile.Friends,
		Enemies:      tornProfile.Enemies,
		Discord:      discordID,
	}

	err = s.repo.StoreProfile(ctx, profile)
	if err != nil {
		return errors.New("error storing profile")
	}

	return nil
}
